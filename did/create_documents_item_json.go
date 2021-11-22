package did

import (
	"github.com/spikeekips/mitum-currency/currency"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type CreateDocumentsItemJSONPacker struct {
	jsonenc.HintedHead
	SM Summary             `json:"summary"`
	DI currency.Big        `json:"documentid"`
	SC string              `json:"signcode"`
	CT string              `json:"content"`
	SZ currency.Big        `json:"size"`
	CI currency.CurrencyID `json:"currency"`
}

func (it BaseCreateDocumentsItem) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(CreateDocumentsItemJSONPacker{
		HintedHead: jsonenc.NewHintedHead(it.Hint()),
		SM:         it.summary,
		DI:         it.documentid,
		SC:         it.signcode,
		CT:         it.content,
		SZ:         it.size,
		CI:         it.cid,
	})
}

type CreateDocumentsItemJSONUnpacker struct {
	SM string                `json:"summary"`
	DI currency.Big          `json:"documentid"`
	SC string                `json:"signcode"`
	CT string                `json:"content"`
	SZ currency.Big          `json:"size"`
	CI string                `json:"currency"`
}

func (it *BaseCreateDocumentsItem) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ht jsonenc.HintedHead
	if err := enc.Unmarshal(b, &ht); err != nil {
		return err
	}

	var ucd CreateDocumentsItemJSONUnpacker
	if err := jsonenc.Unmarshal(b, &ucd); err != nil {
		return err
	}

	return it.unpack(enc, ht.H, ucd.SM, ucd.DI, ucd.SC, ucd.CT, ucd.SZ, ucd.CI)
}
