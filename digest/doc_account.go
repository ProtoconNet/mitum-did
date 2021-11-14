package digest

import (
	"github.com/pkg/errors"
	"github.com/ProtoconNet/mitum-did/did"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/state"
	mongodbstorage "github.com/spikeekips/mitum/storage/mongodb"
	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
)

type AccountDoc struct {
	mongodbstorage.BaseDoc
	address string
	height  base.Height
}

func NewAccountDoc(rs AccountValue, enc encoder.Encoder) (AccountDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, rs, enc)
	if err != nil {
		return AccountDoc{}, err
	}

	return AccountDoc{
		BaseDoc: b,
		address: currency.StateAddressKeyPrefix(rs.ac.Address()),
		height:  rs.height,
	}, nil
}

func (doc AccountDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["address"] = doc.address
	m["height"] = doc.height

	return bsonenc.Marshal(m)
}

type BalanceDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	am currency.Amount
}

// NewBalanceDoc gets the State of Amount
func NewBalanceDoc(st state.State, enc encoder.Encoder) (BalanceDoc, error) {
	am, err := currency.StateBalanceValue(st)
	if err != nil {
		return BalanceDoc{}, errors.Errorf("BalanceDoc needs Amount state: %q", err)
	}

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return BalanceDoc{}, err
	}

	return BalanceDoc{
		BaseDoc: b,
		st:      st,
		am:      am,
	}, nil
}

func (doc BalanceDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}
	address := doc.st.Key()[:len(doc.st.Key())-len(currency.StateKeyBalanceSuffix)-len(doc.am.Currency())-1]
	m["address"] = address
	m["currency"] = doc.am.Currency().String()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}

type DocumentsDoc struct {
	mongodbstorage.BaseDoc
	st state.State
}

// NewDocumentDoc gets the State of DocumentData
func NewDocumentsDoc(st state.State, enc encoder.Encoder) (DocumentsDoc, error) {

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DocumentsDoc{}, err
	}
	return DocumentsDoc{
		BaseDoc: b,
		st:      st,
	}, nil
}

func (doc DocumentsDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}
	address := doc.st.Key()[:len(doc.st.Key())-len(did.StateKeyDocumentsSuffix)]
	m["address"] = address
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}
