package cmds

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/ProtoconNet/mitum-did/did"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

type SummaryFlag struct {
	SM did.Summary
}

func (v *SummaryFlag) UnmarshalText(b []byte) error {
	sm := did.Summary(string(b))
	if err := sm.IsValid(nil); err != nil {
		return err
	}
	v.SM = sm

	return nil
}

func (v *SummaryFlag) String() string {
	return v.SM.String()
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
