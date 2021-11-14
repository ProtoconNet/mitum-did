package digest

import (
	"github.com/ProtoconNet/mitum-did/did"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
)

func (dv *DocumentValue) unpack(enc encoder.Encoder, bdm []byte, height base.Height) error {

	if bdm != nil {
		i, err := did.DecodeDocumentData(bdm, enc)
		if err != nil {
			return err
		}
		dv.doc = i
	}

	dv.height = height

	return nil
}
