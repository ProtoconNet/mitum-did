package blocksign

import (
	"golang.org/x/xerrors"

	"github.com/soonkuk/mitum-data/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	CreateDocumentsFactType = hint.Type("mitum-currency-create-documents-operation-fact")
	CreateDocumentsFactHint = hint.NewHint(CreateDocumentsFactType, "v0.0.1")
	CreateDocumentsType     = hint.Type("mitum-currency-create-documents-operation")
	CreateDocumentsHint     = hint.NewHint(CreateDocumentsType, "v0.0.1")
)

var MaxCreateDocumentsItems uint = 10

type FileDataItem interface {
	FileData() FileData
}

type CreateDocumentsItem interface {
	hint.Hinter
	isvalid.IsValider
	Bytes() []byte
	Keys() currency.Keys
	SignCode() SignCode
	Owner() base.Address
	// Signers() []base.Address
	Address() (base.Address, error)
	Currency() currency.CurrencyID
	Rebuild() CreateDocumentsItem
}

type CreateDocumentsFact struct {
	h      valuehash.Hash
	token  []byte
	sender base.Address
	items  []CreateDocumentsItem
}

func NewCreateDocumentsFact(token []byte, sender base.Address, items []CreateDocumentsItem) CreateDocumentsFact {
	fact := CreateDocumentsFact{
		token:  token,
		sender: sender,
		items:  items,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func (fact CreateDocumentsFact) Hint() hint.Hint {
	return CreateDocumentsFactHint
}

func (fact CreateDocumentsFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact CreateDocumentsFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateDocumentsFact) Bytes() []byte {
	is := make([][]byte, len(fact.items))
	for i := range fact.items {
		is[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.token,
		fact.sender.Bytes(),
		util.ConcatBytesSlice(is...),
	)
}

func (fact CreateDocumentsFact) IsValid([]byte) error {
	if len(fact.token) < 1 {
		return xerrors.Errorf("empty token for CreateDocumentsFact")
	} else if n := len(fact.items); n < 1 {
		return xerrors.Errorf("empty items")
	} else if n > int(MaxCreateDocumentsItems) {
		return xerrors.Errorf("items, %d over max, %d", n, MaxCreateDocumentsItems)
	}

	if err := isvalid.Check([]isvalid.IsValider{
		fact.h,
		fact.sender,
	}, nil, false); err != nil {
		return err
	}

	foundKeys := map[string]struct{}{}
	for i := range fact.items {
		if err := fact.items[i].IsValid(nil); err != nil {
			return err
		}

		it := fact.items[i]
		k := it.Keys().Hash().String()
		if _, found := foundKeys[k]; found {
			return xerrors.Errorf("duplicated acocunt Keys found, %s", k)
		}

		switch a, err := it.Address(); {
		case err != nil:
			return err
		case fact.sender.Equal(a):
			return xerrors.Errorf("target document address is same with sender, %q", fact.sender)
		default:
			foundKeys[k] = struct{}{}
		}
	}

	if !fact.h.Equal(fact.GenerateHash()) {
		return isvalid.InvalidError.Errorf("wrong Fact hash")
	}

	return nil
}

func (fact CreateDocumentsFact) Token() []byte {
	return fact.token
}

func (fact CreateDocumentsFact) Sender() base.Address {
	return fact.sender
}

func (fact CreateDocumentsFact) Items() []CreateDocumentsItem {
	return fact.items
}

func (fact CreateDocumentsFact) Targets() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items))
	for i := range fact.items {
		if a, err := fact.items[i].Address(); err != nil {
			return nil, err
		} else {
			as[i] = a
		}
	}

	return as, nil
}

func (fact CreateDocumentsFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)

	if tas, err := fact.Targets(); err != nil {
		return nil, err
	} else {
		copy(as, tas)
	}

	as[len(fact.items)] = fact.Sender()

	return as, nil
}

func (fact CreateDocumentsFact) Rebulild() CreateDocumentsFact {
	items := make([]CreateDocumentsItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.h = fact.GenerateHash()

	return fact
}

type CreateDocuments struct {
	operation.BaseOperation
	Memo string
}

func NewCreateDocuments(fact CreateDocumentsFact, fs []operation.FactSign, memo string) (CreateDocuments, error) {
	if bo, err := operation.NewBaseOperationFromFact(CreateDocumentsHint, fact, fs); err != nil {
		return CreateDocuments{}, err
	} else {
		op := CreateDocuments{BaseOperation: bo, Memo: memo}

		op.BaseOperation = bo.SetHash(op.GenerateHash())

		return op, nil
	}
}

func (op CreateDocuments) Hint() hint.Hint {
	return CreateDocumentsHint
}

func (op CreateDocuments) IsValid(networkID []byte) error {
	if err := currency.IsValidMemo(op.Memo); err != nil {
		return err
	}

	return operation.IsValidOperation(op, networkID)
}

func (op CreateDocuments) GenerateHash() valuehash.Hash {
	bs := make([][]byte, len(op.Signs())+1)
	for i := range op.Signs() {
		bs[i] = op.Signs()[i].Bytes()
	}

	bs[len(bs)-1] = []byte(op.Memo)

	e := util.ConcatBytesSlice(op.Fact().Hash().Bytes(), util.ConcatBytesSlice(bs...))

	return valuehash.NewSHA256(e)
}

func (op CreateDocuments) AddFactSigns(fs ...operation.FactSign) (operation.FactSignUpdater, error) {
	if o, err := op.BaseOperation.AddFactSigns(fs...); err != nil {
		return nil, err
	} else {
		op.BaseOperation = o.(operation.BaseOperation)
	}

	op.BaseOperation = op.SetHash(op.GenerateHash())

	return op, nil
}