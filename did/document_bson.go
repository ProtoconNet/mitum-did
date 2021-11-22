package did

import (
	"github.com/spikeekips/mitum-currency/currency"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (doc DocumentData) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(doc.Hint()),
		bson.M{
			"documentinfo": doc.info,
			"creator":      doc.creator,
			"content":        doc.content,
			"size":         doc.size,
		}),
	)
}

type DocumentBSONUnpacker struct {
	DI bson.Raw     `bson:"documentinfo"`
	CR bson.Raw     `bson:"creator"`
	CT string       `bson:"content"`
	SZ currency.Big `bson:"size"`
}

func (doc *DocumentData) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var udoc DocumentBSONUnpacker
	if err := enc.Unmarshal(b, &udoc); err != nil {
		return err
	}

	return doc.unpack(enc, udoc.DI, udoc.CR, udoc.CT, udoc.SZ)
}
