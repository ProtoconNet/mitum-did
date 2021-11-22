package did // nolint:dupl

import (
	"github.com/spikeekips/mitum-currency/currency"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateDocumentsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"summary":   it.summary,
				"documentid": it.documentid,
				"signcode":   it.signcode,
				"content":      it.content,
				"size":       it.size,
				"currency":   it.cid,
			}),
	)
}

type CreateDocumentsItemBSONUnpacker struct {
	SM string                `bson:"summary"`
	DI currency.Big          `bson:"documentid"`
	SC string                `bson:"signcode"`
	CT string                `bson:"content"`
	SZ currency.Big          `bson:"size"`
	CI string                `bson:"currency"`
}

func (it *BaseCreateDocumentsItem) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var ht bsonenc.HintedHead
	if err := enc.Unmarshal(b, &ht); err != nil {
		return err
	}

	var ucd CreateDocumentsItemBSONUnpacker
	if err := bson.Unmarshal(b, &ucd); err != nil {
		return err
	}

	return it.unpack(enc, ht.H, ucd.SM, ucd.DI, ucd.SC, ucd.CT, ucd.SZ, ucd.CI)
}
