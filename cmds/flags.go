package cmds

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/ProtoconNet/mitum-did/did"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

type ContentFlag struct {
	CT did.Content
}

func (v *ContentFlag) UnmarshalText(b []byte) error {
	ct := did.Content(string(b))
	if err := ct.IsValid(nil); err != nil {
		return err
	}
	v.CT = ct

	return nil
}

func (v *ContentFlag) String() string {
	return v.CT.String()
}

type DocSignFlag struct {
	AD base.AddressDecoder
	SC string
	sa string
}

func (v *DocSignFlag) UnmarshalText(b []byte) error {

	docSign := strings.SplitN(string(b), ",", 2)
	if len(docSign) != 2 {
		return errors.Errorf(`wrong formatted; "<string address>,<string signcode>"`)
	}

	v.sa = docSign[0]
	hs, err := hint.ParseHintedString(docSign[0])
	if err != nil {
		return err
	}
	v.AD = base.AddressDecoder{HintedString: encoder.NewHintedString(hs.Hint(), hs.Body())}
	v.SC = docSign[1]

	return nil
}

func (v *DocSignFlag) String() string {
	return v.sa
}
