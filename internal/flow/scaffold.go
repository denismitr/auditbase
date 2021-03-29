package flow

import "github.com/pkg/errors"

// Scaffold the the exchange, queues and binding
func (af *MQActionFlow) Scaffold() error {
	if err := af.mq.DeclareExchange(af.cfg.ExchangeName, af.cfg.ExchangeType); err != nil {
		return errors.Wrap(err, "could not scaffold DirectActionExchange on exchage declaration")
	} else {
		af.lg.Debugf("exchange [%s] of type [%s] declared", af.cfg.ExchangeName, af.cfg.ExchangeType)
	}

	if err := af.mq.DeclareQueue(af.cfg.ActionsCreateQueue); err != nil {
		return errors.Wrapf(err, "could not declare [%s] queue", af.cfg.ActionsCreateQueue)
	}

	if err := af.mq.DeclareQueue(af.cfg.ActionsUpdateQueue); err != nil {
		return errors.Wrapf(err, "could not declare [%s] queue", af.cfg.ActionsUpdateQueue)
	}

	if err := af.mq.Bind(af.cfg.ActionsCreateQueue, af.cfg.ExchangeName, af.cfg.ActionsCreateQueue); err != nil {
		return errors.Wrapf(
			err, "could not bind [%s] queue to [%s] exchange with [%s] key",
			af.cfg.ActionsCreateQueue, af.cfg.ExchangeName, af.cfg.ActionsCreateQueue)
	}

	if err := af.mq.Bind(af.cfg.ActionsUpdateQueue, af.cfg.ExchangeName, af.cfg.ActionsUpdateQueue); err != nil {
		return errors.Wrapf(
			err, "could not bind [%s] queue to [%s] exchange with [%s] key",
			af.cfg.ActionsUpdateQueue, af.cfg.ExchangeName, af.cfg.ActionsUpdateQueue)
	}

	return nil
}
