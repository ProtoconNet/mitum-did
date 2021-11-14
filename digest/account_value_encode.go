package digest

import (
	"github.com/pkg/errors"
	"github.com/ProtoconNet/mitum-did/did"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (va *AccountValue) unpack(enc encoder.Encoder, bac []byte, bl []byte, dm []byte, height, previousHeight base.Height) error {
	if bac != nil {
		i, err := currency.DecodeAccount(bac, enc)
		if err != nil {
			return err
		}
		va.ac = i
	}

	hbl, err := enc.DecodeSlice(bl)
	if err != nil {
		return err
	}

	balance := make([]currency.Amount, len(hbl))
	for i := range hbl {
		j, ok := hbl[i].(currency.Amount)
		if !ok {
			return util.WrongTypeError.Errorf("expected currency.Amount, not %T", hbl[i])
		}
		balance[i] = j
	}

	va.balance = balance

	if hinter, err := enc.Decode(dm); err != nil {
		return err
	} else if k, ok := hinter.(did.DocumentInventory); !ok {
		return errors.Errorf("not DocumentInventory: %T", hinter)
	} else {
		va.document = k
	}

	va.height = height
	va.previousHeight = previousHeight

	return nil
}
