package receiver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/denismitr/auditbase/internal/cache"
	"github.com/denismitr/auditbase/internal/flow"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/utils"
	"github.com/denismitr/auditbase/internal/utils/clock"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("invalid input")
var ErrActionAlreadyProcessed = errors.New("action already processed")
var ErrDataPipelineFailed = errors.New("data pipelined could not accept the new action")

type Receiver struct {
	lg    logger.Logger
	clock clock.Clock
	af    flow.ActionFlow
	uuid4 utils.UUID4Generator
	c     cache.Cacher
}

func New(lg logger.Logger, cl clock.Clock, af flow.ActionFlow, uuid4 utils.UUID4Generator, c cache.Cacher) *Receiver {
	return &Receiver{
		lg: lg,
		clock: cl,
		af: af,
		uuid4: uuid4,
		c: c,
	}
}

type Reg struct {
	Hash         string
	UID          string
	RegisteredAt time.Time
}

func (rc *Receiver) ReceiveOneForUpdate(r io.Reader) (*Reg, error) {
	b, err := readBytes(r)
	if err != nil {
		return nil, err
	}

	hash := createHash(b)

	found, err := rc.c.Has(hash)
	if err != nil {
		rc.lg.Error(errors.Wrap(err, "receiver cache failed"))
	}

	if found {
		return nil, ErrActionAlreadyProcessed
	}

	updateAction, err := rc.createUpdateAction(b, hash)
	if err != nil {
		return nil, err
	}

	if err := rc.c.CreateKey(hash, 5*time.Minute); err != nil {
		rc.lg.Error(errors.Wrap(err, "receiver cache failed"))
	}

	if err := rc.af.SendUpdateAction(updateAction); err != nil {
		return nil, errors.Wrap(ErrDataPipelineFailed, err.Error())
	}

	return &Reg{
		Hash: updateAction.Hash,
		UID:  updateAction.UID,
		RegisteredAt: updateAction.RegisteredAt,
	}, nil
}

func (rc *Receiver) ReceiveOneForCreate(r io.Reader) (*Reg, error) {
	b, err := readBytes(r)
	if err != nil {
		return nil, err
	}

	hash := createHash(b)

	found, err := rc.c.Has(hash)
	if err != nil {
		rc.lg.Error(errors.Wrap(err, "receiver cache failed"))
	}

	if found {
		return nil, ErrActionAlreadyProcessed
	}

	newAction, err := rc.createNewAction(b, hash)
	if err != nil {
		return nil, err
	}

	if err := rc.c.CreateKey(hash, 5*time.Minute); err != nil {
		rc.lg.Error(errors.Wrap(err, "receiver cache failed"))
	}

	if err := rc.af.SendNewAction(newAction); err != nil {
		return nil, errors.Wrap(ErrDataPipelineFailed, err.Error())
	}

	return &Reg{
		Hash: newAction.Hash,
		UID:  newAction.UID,
		RegisteredAt: newAction.RegisteredAt,
	}, nil
}

func readBytes(r io.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidInput, "could not read incoming action payload", err.Error())
	}

	if b == nil || len(b) == 0 {
		return nil, errors.Wrapf(ErrInvalidInput, "empty body of incoming action payload")
	}
	return b, nil
}

func (rc *Receiver) createNewAction(in []byte, hash string) (*model.NewAction, error) {
	newAction := new(model.NewAction)
	if err := json.Unmarshal(in, newAction); err != nil {
		return nil, errors.Wrapf(ErrInvalidInput, "could not parse incoming action payload", err.Error())
	}

	if errorBag := newAction.Validate(); errorBag.NotEmpty() {
		return nil, errorBag // fixme
	}

	newAction.Hash = hash
	newAction.RegisteredAt = rc.clock.CurrentTime()
	if newAction.UID == "" {
		newAction.UID = rc.uuid4.Generate()
	}

	return newAction, nil
}

func (rc *Receiver) createUpdateAction(in []byte, hash string) (*model.UpdateAction, error) {
	updateAction := new(model.UpdateAction)
	if err := json.Unmarshal(in, updateAction); err != nil {
		return nil, errors.Wrapf(ErrInvalidInput, "could not parse incoming action payload", err.Error())
	}

	if errorBag := updateAction.Validate(); errorBag.NotEmpty() {
		return nil, errorBag
	}

	updateAction.Hash = hash
	updateAction.RegisteredAt = rc.clock.CurrentTime()
	if updateAction.UID == "" {
		updateAction.UID = rc.uuid4.Generate()
	}

	return updateAction, nil
}

func createHash(in []byte) string {
	hash := sha256.Sum256(in)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}
