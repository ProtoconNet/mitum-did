package did

import (
	"testing"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/key"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/prprocessor"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util"
	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testCreateDocumentsOperation struct {
	baseTestOperationProcessor
}

func (t *testCreateDocumentsOperation) processor(cp *currency.CurrencyPool, pool *storage.Statepool) prprocessor.OperationProcessor {
	copr, err := NewOperationProcessor(nil).
		SetProcessor(CreateDocuments{}, NewCreateDocumentsProcessor(cp))
	t.NoError(err)

	if pool == nil {
		return copr
	}

	return copr.New(pool)
}

func (t *testCreateDocumentsOperation) newOperation(sender base.Address, items []CreateDocumentsItem, pks []key.Privatekey) CreateDocuments {
	token := util.UUID().Bytes()
	fact := NewCreateDocumentsFact(token, sender, items)

	var fs []operation.FactSign
	for _, pk := range pks {
		sig, err := operation.NewFactSignature(pk, fact, nil)
		if err != nil {
			panic(err)
		}

		fs = append(fs, operation.NewBaseFactSign(pk.Publickey(), sig))
	}

	cd, err := NewCreateDocuments(fact, fs, "")
	if err != nil {
		panic(err)
	}

	err = cd.IsValid(nil)
	if err != nil {
		panic(err)
	}

	return cd
}

func (t *testCreateDocumentsOperation) TestNormalCase() {
	cid := currency.CurrencyID("SHOWME")

	// sender initial balance
	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}

	// sender account
	sa, st0 := t.newAccount(true, balance)
	// signer account
	sga, st1 := t.newAccount(true, balance)

	pool, _ := t.statepool(st0, st1)

	fee := currency.NewBig(1)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	// filedata
	ct := Content("ABCD")
	documentid := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			ct,
			documentid,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	t.NoError(opr.Process(cd))

	// check updated state
	// new documents state
	var ns state.State
	// new document data state
	var nds state.State
	// sender balance state
	var sb state.State
	for _, stu := range pool.Updates() {
		if currency.IsStateBalanceKey(stu.Key()) {
			st := stu.GetState()

			i, err := currency.StateBalanceValue(st)
			t.NoError(err)

			if st.Key() == currency.StateKeyBalance(sa.Address, i.Currency()) {
				sb = st
			} else {
				continue
			}
		} else if (IsStateDocumentsKey(stu.Key())) && (stu.Key() == StateKeyDocuments(sa.Address)) {
			ns = stu.GetState()
		} else if (IsStateDocumentDataKey(stu.Key())) && (stu.Key() == StateKeyDocumentData(DocId(documentid))) {
			nds = stu.GetState()
		}
	}

	t.NotNil(sb)

	sba, _ := currency.StateBalanceValue(sb)
	t.True(sba.Big().Equal(balance[0].Big().Sub(fee)))

	t.Equal(fee, sb.(currency.AmountState).Fee())

	ndd, _ := StateDocumentDataValue(nds)
	t.True(ndd.Content().Equal(ct))
	t.True(ndd.Creator().Equal(sa.Address))

	ndinv, _ := StateDocumentsValue(ns)
	t.True(ndinv.Documents()[0].Content().Equal(ct))
}

func (t *testCreateDocumentsOperation) TestSignerAccountsNotExist() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}
	sa, st := t.newAccount(true, balance)
	sga, _ := t.newAccount(false, nil)

	pool, _ := t.statepool(st)
	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	// filedata
	ct := Content("ABCD")
	documentid := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			ct,
			documentid,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	err := opr.Process(cd)

	var oper operation.ReasonError
	t.True(xerrors.As(err, &oper))
	t.Contains(err.Error(), "signer account not found")
}

func (t *testCreateDocumentsOperation) TestDocumentAlreadyExists() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}

	sa0, st0 := t.newAccount(true, balance)
	sa1, st1 := t.newAccount(true, balance)

	content := Content("ABCD")
	documentid := currency.NewBig(0)
	info := DocInfo{
		idx:      documentid,
		content: content,
	}
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)

	doc := NewDocumentData(info, sa1.Address, signcode0, title, size, []DocSign{})

	nds := t.newStateDocumentData(doc)

	pool, _ := t.statepool(st0, st1, []state.State{nds})

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa1.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa0.Address, feeer)))

	opr := t.processor(cp, pool)

	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid,
			signcode0,
			title,
			size,
			[]base.Address{},
			[]string{},
			cid,
		),
	}
	cd := t.newOperation(sa0.Address, items, sa0.Privs())

	err := opr.Process(cd)

	var oper operation.ReasonError
	t.True(xerrors.As(err, &oper))
	t.Contains(err.Error(), "documentid already registered")
}

func (t *testCreateDocumentsOperation) TestSameSenders() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}

	sa, sta0 := t.newAccount(true, balance)
	sga, sta1 := t.newAccount(true, balance)

	pool, _ := t.statepool(sta0, sta1)

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	documentid1 := currency.NewBig(1)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items0 := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd0 := t.newOperation(sa.Address, items0, sa.Privs())
	t.NoError(opr.Process(cd0))

	items1 := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid1,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd1 := t.newOperation(sa.Address, items1, sa.Privs())

	err := opr.Process(cd1)
	t.Contains(err.Error(), "violates only one sender")
}

func (t *testCreateDocumentsOperation) TestSameSendersWithInvalidOperation() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}

	sa, sta0 := t.newAccount(true, balance)
	sga, sta1 := t.newAccount(true, balance)

	pool, _ := t.statepool(sta0, sta1)

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	documentid1 := currency.NewBig(1)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"

	// insert invalid operation, under threshold signing. It can not be counted
	// to sender checking.
	{
		items := []CreateDocumentsItem{
			NewCreateDocumentsItemSingleFile(
				content,
				documentid0,
				signcode0,
				title,
				size,
				[]base.Address{sga.Address},
				[]string{signcode1},
				cid,
			),
		}
		cd := t.newOperation(sa.Address, items, []key.Privatekey{key.MustNewBTCPrivatekey()})
		err := opr.Process(cd)

		var oper operation.ReasonError
		t.True(xerrors.As(err, &oper))
	}

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items0 := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd0 := t.newOperation(sa.Address, items0, sa.Privs())
	t.NoError(opr.Process(cd0))

	items1 := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid1,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd1 := t.newOperation(sa.Address, items1, sa.Privs())

	err := opr.Process(cd1)
	t.Contains(err.Error(), "violates only one sender")
}

func (t *testCreateDocumentsOperation) TestSignerSameWithOwner() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}

	sa, sta := t.newAccount(true, balance)
	sga := sa

	pool, _ := t.statepool(sta)

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items0 := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd := t.newOperation(sa.Address, items0, sa.Privs())
	err := opr.Process(cd)

	err = opr.Process(cd)
	t.Contains(err.Error(), "signer account is same with document creator")
}

func (t *testCreateDocumentsOperation) TestDuplicatedSigner() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid),
	}

	sa, sta0 := t.newAccount(true, balance)
	sga, sta1 := t.newAccount(true, balance)

	pool, _ := t.statepool(sta0, sta1)

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"
	signcode2 := "user2"

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items0 := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address, sga.Address},
			[]string{signcode1, signcode2},
			cid,
		),
	}
	cd := t.newOperation(sa.Address, items0, sa.Privs())
	err := opr.Process(cd)

	err = opr.Process(cd)
	t.Contains(err.Error(), "duplicated signer")
}

func (t *testCreateDocumentsOperation) TestMultipleItemsWithFee() {
	cid0 := currency.CurrencyID("SHOWME")
	cid1 := currency.CurrencyID("FINDME")

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid0),
		currency.NewAmount(currency.NewBig(33), cid1),
	}

	sa, st := t.newAccount(true, balance)

	pool, _ := t.statepool(st)

	fee0 := currency.NewBig(1)
	fee1 := currency.NewBig(2)
	feeer0 := currency.NewFixedFeeer(sa.Address, fee0)
	feeer1 := currency.NewFixedFeeer(sa.Address, fee1)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid0, currency.NewBig(99), sa.Address, feeer0)))
	t.NoError(cp.Set(t.newCurrencyDesignState(cid1, currency.NewBig(99), sa.Address, feeer1)))

	opr := t.processor(cp, pool)

	content0 := Content("ABCD")
	content1 := Content("EFGH")
	documentid0 := currency.NewBig(0)
	documentid1 := currency.NewBig(1)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)

	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content0,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{},
			[]string{},
			cid0,
		),
		NewCreateDocumentsItemSingleFile(
			content1,
			documentid1,
			signcode0,
			title,
			size,
			[]base.Address{},
			[]string{},
			cid1,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	t.NoError(opr.Process(cd))

	// check updated state
	// new documents state
	var ns state.State
	// new document data state
	var dds0, dds1 state.State
	// sender balance state
	sb := map[currency.CurrencyID]state.State{}
	for _, stu := range pool.Updates() {
		if currency.IsStateBalanceKey(stu.Key()) {
			st := stu.GetState()

			i, err := currency.StateBalanceValue(st)
			t.NoError(err)

			if st.Key() == currency.StateKeyBalance(sa.Address, i.Currency()) {
				sb[i.Currency()] = st
			} else {
				continue
			}
		} else if IsStateDocumentsKey(stu.Key()) {
			if stu.Key() == StateKeyDocuments(sa.Address) {
				ns = stu.GetState()
			}
		} else if IsStateDocumentDataKey(stu.Key()) {
			if stu.Key() == StateKeyDocumentData(DocId(documentid0)) {
				dds0 = stu.GetState()
			} else if stu.Key() == StateKeyDocumentData(DocId(documentid1)) {
				dds1 = stu.GetState()
			}
		}
	}

	udinv := ns.Value().Interface().(DocumentInventory)

	t.True(content0.Equal(udinv.Documents()[0].Content()))
	t.True(content1.Equal(udinv.Documents()[1].Content()))

	t.Equal(len(balance), len(sb))

	sba0, _ := currency.StateBalanceValue(sb[cid0])
	t.True(sba0.Big().Equal(balance[0].Big().Sub(fee0)))

	sba1, _ := currency.StateBalanceValue(sb[cid1])
	t.True(sba1.Big().Equal(balance[1].Big().Sub(fee1)))

	t.Equal(fee0, sb[cid0].(currency.AmountState).Fee())
	t.Equal(fee1, sb[cid1].(currency.AmountState).Fee())

	ndd0, _ := StateDocumentDataValue(dds0)
	t.True(ndd0.Content().Equal(content0))
	t.True(ndd0.Creator().Equal(sa.Address))

	ndd1, _ := StateDocumentDataValue(dds1)
	t.True(ndd1.Content().Equal(content1))
	t.True(ndd1.Creator().Equal(sa.Address))
}

func (t *testCreateDocumentsOperation) TestInSufficientBalanceForFee() {
	// currency id
	cid := currency.CurrencyID("SHOWME")
	// sender balance
	senderBalance := int64(3)
	// fee amount
	feeAmount := int64(4)

	// sender initial balance
	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(senderBalance), cid),
	}

	// sender account
	sa, st := t.newAccount(true, balance)

	pool, _ := t.statepool(st)

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	// filedata
	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{},
			[]string{},
			cid,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	err := opr.Process(cd)

	var oper operation.ReasonError
	t.True(xerrors.As(err, &oper))
	t.Contains(err.Error(), "insufficient balance")
}

func (t *testCreateDocumentsOperation) TestUnknownCurrencyID() {
	// currency id of network
	cid0 := currency.CurrencyID("SHOWME")
	// currency id used in operation
	cid1 := currency.CurrencyID("FINDME")

	// sender initial balance
	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid0),
	}

	// sender account
	sa, st := t.newAccount(true, balance)

	pool, _ := t.statepool(st)

	fee := currency.NewBig(2)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid0, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	// filedata
	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{},
			[]string{},
			cid1,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	err := opr.Process(cd)

	var oper operation.ReasonError
	t.True(xerrors.As(err, &oper))
	t.Contains(err.Error(), "unknown currency id found")
}

func (t *testCreateDocumentsOperation) TestEmptyCurrency() {
	cid0 := currency.CurrencyID("FINDME")
	cid1 := currency.CurrencyID("SHOWME")
	feeAmount := int64(1)

	balance := []currency.Amount{
		currency.NewAmount(currency.NewBig(33), cid0),
	}
	sa, st := t.newAccount(true, balance)

	pool, _ := t.statepool(st)
	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid0, currency.NewBig(99), sa.Address, feeer)))
	t.NoError(cp.Set(t.newCurrencyDesignState(cid1, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	// filedata
	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{},
			[]string{},
			cid1,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	err := opr.Process(cd)

	var oper operation.ReasonError
	t.True(xerrors.As(err, &oper))
	t.Contains(err.Error(), "currency of holder does not exist")
}

func (t *testCreateDocumentsOperation) TestSenderBalanceNotExist() {
	cid := currency.CurrencyID("FINDME")
	feeAmount := int64(1)

	sa, st1 := t.newAccount(true, nil)
	sga, st2 := t.newAccount(true, nil)

	pool, _ := t.statepool(st1, st2)

	fee := currency.NewBig(feeAmount)
	feeer := currency.NewFixedFeeer(sa.Address, fee)

	cp := currency.NewCurrencyPool()
	t.NoError(cp.Set(t.newCurrencyDesignState(cid, currency.NewBig(99), sa.Address, feeer)))

	opr := t.processor(cp, pool)

	// filedata
	content := Content("ABCD")
	documentid0 := currency.NewBig(0)
	signcode0 := "user0"
	title := "title01"
	size := currency.NewBig(555)
	signcode1 := "user1"

	// create document of na(document account) with oa(owner) which is sent from sa(sender)
	items := []CreateDocumentsItem{
		NewCreateDocumentsItemSingleFile(
			content,
			documentid0,
			signcode0,
			title,
			size,
			[]base.Address{sga.Address},
			[]string{signcode1},
			cid,
		),
	}
	cd := t.newOperation(sa.Address, items, sa.Privs())

	err := opr.Process(cd)

	var oper operation.ReasonError
	t.True(xerrors.As(err, &oper))
	t.Contains(err.Error(), "currency of holder does not exist")
}

func TestCreateDocumentsOperation(t *testing.T) {
	suite.Run(t, new(testCreateDocumentsOperation))
}
