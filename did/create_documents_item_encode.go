package did

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (it *BaseCreateDocumentsItem) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	bsm string,
	bdi currency.Big,
	bsc string,
	bct string,
	bsz currency.Big,
	scid string,

) error {
	it.hint = ht

	it.summary = Summary(bsm)
	it.documentid = bdi
	it.signcode = bsc
	it.content = bct
	it.size = bsz
	it.cid = currency.CurrencyID(scid)

	return nil
}
