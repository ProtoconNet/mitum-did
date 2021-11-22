package did

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

type BaseCreateDocumentsItem struct {
	hint       hint.Hint
	summary    Summary
	documentid currency.Big
	signcode   string //creator signcode
	content      string
	size       currency.Big
	cid        currency.CurrencyID
}

func NewBaseCreateDocumentsItem(ht hint.Hint,
	summary Summary,
	documentid currency.Big,
	signcode, content string,
	size currency.Big,
	cid currency.CurrencyID) BaseCreateDocumentsItem {
	return BaseCreateDocumentsItem{
		hint:       ht,
		summary:   summary,
		documentid: documentid,
		signcode:   signcode,
		content:      content,
		size:       size,
		cid:        cid,
	}
}

func (it BaseCreateDocumentsItem) Hint() hint.Hint {
	return it.hint
}

func (it BaseCreateDocumentsItem) Bytes() []byte {
	bs := make([][]byte, 6)
	bs[0] = it.summary.Bytes()
	bs[1] = it.documentid.Bytes()
	bs[2] = []byte(it.signcode)
	bs[3] = []byte(it.content)
	bs[4] = it.size.Bytes()
	bs[5] = it.cid.Bytes()

	return util.ConcatBytesSlice(bs...)
}

func (it BaseCreateDocumentsItem) IsValid([]byte) error {
	if len(it.summary) < 1 {
		return errors.Errorf("empty summary")
	}
	if (it.documentid == currency.Big{}) {
		return errors.Errorf("empty documentid")
	}
	if !it.documentid.OverZero() {
		return errors.Errorf("documentid is negative number")
	}
	if len(it.signcode) < 1 {
		return errors.Errorf("empty creator signcode")
	}
	if err := it.cid.IsValid(nil); err != nil {
		return err
	}
	return nil
}

// Summary return BaseCreateDocumetsItem's owner address.
func (it BaseCreateDocumentsItem) Summary() Summary {
	return it.summary
}

func (it BaseCreateDocumentsItem) DocumentId() currency.Big {
	return it.documentid
}

func (it BaseCreateDocumentsItem) Signcode() string {
	return it.signcode
}

func (it BaseCreateDocumentsItem) Content() string {
	return it.content
}

func (it BaseCreateDocumentsItem) Size() currency.Big {
	return it.size
}

// FileData return BaseCreateDocumentsItem's fileData.
func (it BaseCreateDocumentsItem) Currency() currency.CurrencyID {
	return it.cid
}

func (it BaseCreateDocumentsItem) Rebuild() CreateDocumentsItem {
	return it
}
