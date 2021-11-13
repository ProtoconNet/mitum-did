package blocksign

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact *CreateDocumentsFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	tk []byte,
	bSender base.AddressDecoder,
	bits []byte,
) error {
	sender, err := bSender.Encode(enc)
	if err != nil {
		return err
	}

	hits, err := enc.DecodeSlice(bits)
	if err != nil {
		return err
	}

	its := make([]CreateDocumentsItem, len(hits))
	for i := range hits {
		j, ok := hits[i].(CreateDocumentsItem)
		if !ok {
			return util.WrongTypeError.Errorf("expected CreateDocumentsItem, not %T", hits[i])
		}

		its[i] = j
	}

	fact.h = h
	fact.token = tk
	fact.sender = sender
	fact.items = its

	return nil
}
