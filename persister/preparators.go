package persister

import (
	"context"
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const ErrTimeout = errtype.StringError("timeout occurred")

func prepareService(
	ctx context.Context,
	ms model.Microservice,
	rememberFunc cache.RememberFunc,
	factory model.RepositoryFactory,
) (*model.Microservice, error) {
	serviceCh := make(chan *model.Microservice)
	errCh := make(chan error)

	go func() {
		serviceCacheKey := model.MicroserviceItemCacheKey(ms.Name)
		service := new(model.Microservice)

		if err := rememberFunc(serviceCacheKey, 3*time.Minute, service, func() (interface{}, error) {
			v, err := factory.Microservices().FirstOrCreateByName(ctx, ms.Name)
			if err != nil {
				return nil, errors.Wrap(err, "service cache error")
			}

			return v, nil
		}); err != nil {
			errCh <- err
			return
		}

		serviceCh <- service
	}()

	select {
	case err := <-errCh:
		return nil, err
	case service := <-serviceCh:
		return service, nil
	case <-ctx.Done():
		return nil, errors.Wrapf(ErrTimeout, "could not prepare a service %s", ms.Name)
	}
}

func prepareEntity(
	ctx context.Context,
	service *model.Microservice,
	e model.Entity,
	rememberFunc cache.RememberFunc,
	factory model.RepositoryFactory,
) (*model.Entity, error) {
	entityCh := make(chan *model.Entity)
	errCh := make(chan error)

	go func() {
		entity := new(model.Entity)
		entityCacheKey := model.EntityItemCacheKey(e.Name, service)

		if err := rememberFunc(entityCacheKey, 5*time.Minute, entity, func() (interface{}, error) {
			v, err := factory.Entities().FirstOrCreateByNameAndService(e.Name, service)
			if err != nil {
				return nil, err
			}
			return v, nil
		}); err != nil {
			errCh <- err
			return
		}

		entityCh <- entity
	}()

	select {
	case err := <-errCh:
		return nil, err
	case entity := <-entityCh:
		return entity, nil
	case <-ctx.Done():
		return nil, errors.Wrapf(ErrTimeout, "could not prepare an entity %s", e.Name)
	}
}

func prepareEntityProperties(
	ctx context.Context,
	propertyNames []string,
	factory model.RepositoryFactory,
	entity model.Entity,
) (map[string]string, error) {
	var resultMu sync.Mutex
	errCh := make(chan error)
	result := make(map[string]string)

	var wg sync.WaitGroup

	for _, name := range propertyNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			id, err := factory.Properties().GetIDOrCreate(name, entity.ID)
			if err != nil {
				errCh <- err
				return
			}

			if !validator.IsUUID4(id) {
				errCh <- errors.Errorf("property id %s invalid", id)
				return
			}

			resultMu.Lock()
			result[id] = name
			resultMu.Unlock()
		}(name)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		resultMu.Lock()
		defer resultMu.Unlock()
		return result, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ErrTimeout
	}
}