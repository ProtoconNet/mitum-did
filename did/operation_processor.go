package did

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/prprocessor"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/logging"
	"github.com/spikeekips/mitum/util/valuehash"
)

type GetNewProcessor func(state.Processor) (state.Processor, error)

type DuplicationType string

const (
	DuplicationTypeSender   DuplicationType = "sender"
	DuplicationTypeCurrency DuplicationType = "currency"
)

type OperationProcessor struct {
	sync.RWMutex
	*logging.Logging
	processorHintSet     *hint.Hintmap
	cp                   *currency.CurrencyPool
	pool                 *storage.Statepool
	fee                  map[currency.CurrencyID]currency.Big
	amountPool           map[string]currency.AmountState
	duplicated           map[string]DuplicationType
	duplicatedNewAddress map[string]struct{}
}

func NewOperationProcessor(cp *currency.CurrencyPool) *OperationProcessor {
	return &OperationProcessor{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "mitum-currency-operations-processor")
		}),
		processorHintSet: hint.NewHintmap(),
		cp:               cp,
	}
}

func (opr *OperationProcessor) New(pool *storage.Statepool) prprocessor.OperationProcessor {
	return &OperationProcessor{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "mitum-currency-operations-processor")
		}),
		processorHintSet:     opr.processorHintSet,
		cp:                   opr.cp,
		pool:                 pool,
		fee:                  map[currency.CurrencyID]currency.Big{},
		amountPool:           map[string]currency.AmountState{},
		duplicated:           map[string]DuplicationType{},
		duplicatedNewAddress: map[string]struct{}{},
	}
}

func (opr *OperationProcessor) SetProcessor(
	hinter hint.Hinter,
	newProcessor currency.GetNewProcessor,
) (prprocessor.OperationProcessor, error) {
	if err := opr.processorHintSet.Add(hinter, newProcessor); err != nil {
		return nil, err
	}
	return opr, nil
}

func (opr *OperationProcessor) setState(op valuehash.Hash, sts ...state.State) error {
	opr.Lock()
	defer opr.Unlock()

	for i := range sts {
		if t, ok := sts[i].(currency.AmountState); ok {
			if t.Fee().OverZero() {
				f := currency.ZeroBig
				if i, found := opr.fee[t.Currency()]; found {
					f = i
				}

				opr.fee[t.Currency()] = f.Add(t.Fee())
			}
		}
	}

	return opr.pool.Set(op, sts...)
}

func (opr *OperationProcessor) PreProcess(op state.Processor) (state.Processor, error) {
	var sp state.Processor
	switch i, known, err := opr.getNewProcessor(op); {
	case err != nil:
		return nil, operation.NewBaseReasonErrorFromError(err)
	case !known:
		return op, nil
	default:
		sp = i
	}

	pop, err := sp.(state.PreProcessor).PreProcess(opr.pool.Get, opr.setState)
	if err != nil {
		return nil, err
	}

	if err := opr.checkDuplication(op); err != nil {
		return nil, operation.NewBaseReasonError("duplication found: %w", err)
	}

	return pop, nil
}

func (opr *OperationProcessor) Process(op state.Processor) error {
	switch op.(type) {
	case *currency.TransfersProcessor,
		*currency.CreateAccountsProcessor,
		*currency.KeyUpdaterProcessor,
		*currency.CurrencyRegisterProcessor,
		*currency.CurrencyPolicyUpdaterProcessor,
		*CreateDocumentsProcessor,
		*SignDocumentsProcessor:
		return opr.process(op)
	case currency.Transfers, currency.CreateAccounts, currency.KeyUpdater, currency.CurrencyRegister, currency.CurrencyPolicyUpdater, CreateDocuments, SignDocuments:
		pr, err := opr.PreProcess(op)
		if err != nil {
			return err
		}
		return opr.process(pr)
	default:
		return op.Process(opr.pool.Get, opr.pool.Set)
	}
}

func (opr *OperationProcessor) process(op state.Processor) error {
	var sp state.Processor

	switch t := op.(type) {
	case *currency.TransfersProcessor:
		sp = t
	case *currency.CreateAccountsProcessor:
		sp = t
	case *currency.KeyUpdaterProcessor:
		sp = t
	case *CreateDocumentsProcessor:
		sp = t
	case *SignDocumentsProcessor:
		sp = t
	default:
		return op.Process(opr.pool.Get, opr.pool.Set)
	}

	return sp.Process(opr.pool.Get, opr.setState)
}

func (opr *OperationProcessor) checkDuplication(op state.Processor) error {
	opr.Lock()
	defer opr.Unlock()

	var did string
	var didtype DuplicationType
	var newAddresses []base.Address

	switch t := op.(type) {
	case currency.Transfers:
		did = t.Fact().(currency.TransfersFact).Sender().String()
		didtype = DuplicationTypeSender
	case currency.CreateAccounts:
		fact := t.Fact().(currency.CreateAccountsFact)
		as, err := fact.Targets()
		if err != nil {
			return errors.Errorf("failed to get Addresses")
		}
		newAddresses = as
		did = fact.Sender().String()
		didtype = DuplicationTypeSender
	case currency.KeyUpdater:
		did = t.Fact().(currency.KeyUpdaterFact).Target().String()
		didtype = DuplicationTypeSender
	case currency.CurrencyRegister:
		did = t.Fact().(currency.CurrencyRegisterFact).Currency().Currency().String()
		didtype = DuplicationTypeCurrency
	case currency.CurrencyPolicyUpdater:
		did = t.Fact().(currency.CurrencyPolicyUpdaterFact).Currency().String()
		didtype = DuplicationTypeCurrency
	case CreateDocuments:
		did = t.Fact().(CreateDocumentsFact).Sender().String()
		didtype = DuplicationTypeSender
	case SignDocuments:
		did = t.Fact().(SignDocumentsFact).Sender().String()
		didtype = DuplicationTypeSender
	default:
		return nil
	}

	if len(did) > 0 {
		if _, found := opr.duplicated[did]; found {
			switch didtype {
			case DuplicationTypeSender:
				return errors.Errorf("violates only one sender in proposal")
			case DuplicationTypeCurrency:
				return errors.Errorf("duplicated currency id, %q found in proposal", did)
			default:
				return errors.Errorf("violates duplication in proposal")
			}
		}

		opr.duplicated[did] = didtype
	}

	if len(newAddresses) > 0 {
		if err := opr.checkNewAddressDuplication(newAddresses); err != nil {
			return err
		}
	}

	return nil
}

func (opr *OperationProcessor) checkNewAddressDuplication(as []base.Address) error {
	for i := range as {
		if _, found := opr.duplicatedNewAddress[as[i].String()]; found {
			return errors.Errorf("new address already processed")
		}
	}

	for i := range as {
		opr.duplicatedNewAddress[as[i].String()] = struct{}{}
	}

	return nil
}

func (opr *OperationProcessor) Close() error {
	opr.RLock()
	defer opr.RUnlock()

	if opr.cp != nil && len(opr.fee) > 0 {
		op := currency.NewFeeOperation(currency.NewFeeOperationFact(opr.pool.Height(), opr.fee))

		pr := currency.NewFeeOperationProcessor(opr.cp, op)
		if err := pr.Process(opr.pool.Get, opr.pool.Set); err != nil {
			return err
		}
		opr.pool.AddOperations(op)
	}

	return nil
}

func (opr *OperationProcessor) Cancel() error {
	opr.RLock()
	defer opr.RUnlock()

	return nil
}

func (opr *OperationProcessor) getNewProcessor(op state.Processor) (state.Processor, bool, error) {
	switch i, err := opr.getNewProcessorFromHintset(op); {
	case err != nil:
		return nil, false, err
	case i != nil:
		return i, true, nil
	}

	switch t := op.(type) {
	case currency.Transfers,
		currency.CreateAccounts,
		currency.KeyUpdater,
		currency.CurrencyRegister,
		currency.CurrencyPolicyUpdater,
		CreateDocuments,
		SignDocuments:
		return nil, false, errors.Errorf("%T needs SetProcessor", t)
	default:
		return op, false, nil
	}
}

func (opr *OperationProcessor) getNewProcessorFromHintset(op state.Processor) (state.Processor, error) {
	var f currency.GetNewProcessor
	if hinter, ok := op.(hint.Hinter); !ok {
		return nil, nil
	} else if i, err := opr.processorHintSet.Compatible(hinter); err != nil {
		if errors.Is(err, util.NotFoundError) {
			return nil, nil
		}

		return nil, err
	} else if j, ok := i.(currency.GetNewProcessor); !ok {
		return nil, errors.Errorf("invalid GetNewProcessor func, %q", i)
	} else {
		f = j
	}

	return f(op)
}
