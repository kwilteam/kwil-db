package host

import (
	"context"
	"sync"

	cm "kwil/x/common"
	cfg "kwil/x/common/host/config"
	types "kwil/x/common/host/service"
)

var status int8
var servicesById = make(map[int32]types.ServiceFactory)
var servicesByName = make(map[string]types.ServiceFactory)
var mu = sync.RWMutex{}
var wg = sync.WaitGroup{}

var fnByIdUnsafe = func(id int32) (types.Service, error) {
	return getServiceByIdUnsafe[types.Service](id)
}

var fnByNameUnsafe = func(name string) (types.Service, error) {
	return getServiceByNameUnsafe[types.Service](name)
}

// IsRegistered returns true if the service is registered and the host is running.
func IsRegistered[T any](id int32) bool {
	mu.RLock()
	defer mu.RUnlock()

	if status != 0 {
		return false
	}

	if _, ok := servicesById[id]; !ok {
		return false
	}

	return true
}

func GetServiceByName[T any](name string) (service T, e error) {
	mu.RLock()
	defer mu.RUnlock()

	if status == 0 {
		e = ErrHostNotRunning
		return
	}

	return getServiceByNameUnsafe[T](name)
}

func GetServiceById[T any](id int32) (service T, e error) {
	mu.RLock()
	defer mu.RUnlock()

	if status == 0 {
		e = ErrHostNotRunning
		return
	}

	return getServiceByIdUnsafe[T](id)
}

func RegisterService(name string, factory types.ServiceFactory) (identity types.ServiceIdentity, err error) {
	identity, err = types.NewServiceIdentity(name)
	if err != nil {
		return types.ServiceIdentity{}, err
	}

	if identity.Name() == "" || identity.Id() <= 0 {
		return types.ServiceIdentity{}, ErrInvalidServiceIdentity
	}

	mu.Lock()
	defer mu.Unlock()

	if status != 0 {
		err = ErrHostRegistrationClosed
		return
	}

	if _, exists := servicesById[identity.Id()]; exists {
		return
	}

	if _, exists := servicesByName[identity.Name()]; exists {
		return
	}

	var o = sync.Once{}
	var s types.Service
	var fn = func() types.Service {
		o.Do(func() {
			s = factory()
		})

		return s
	}

	servicesById[identity.Id()] = fn
	servicesByName[identity.Name()] = fn

	wg.Add(1)

	return
}

func Start(c context.Context, config cfg.Config) error {
	err := doStartup(config)
	if err != nil {
		return err
	}

	ctx := newServiceContext(c, fnByIdUnsafe, fnByNameUnsafe)

	// Initialize each service (service can retrieve other services via GetService at this stage)
	for _, s := range servicesByName {
		var s2 = s()
		if s2 != nil {
			v, ok := s2.(types.Initializeable)
			if ok {
				v.Initialize(ctx)
			}
		}
	}

	// Start bacground services
	for _, s := range servicesByName {
		v, ok := s().(types.BackgroundService)
		if !ok {
			continue
		}

		if err := v.Start(ctx); err != nil {
			shutdownUnsafe()
			status = 86
			return err
		}
	}

	return nil
}

func Shutdown() error {
	mu.Lock()
	defer mu.Unlock()

	if status == 86 {
		return ErrHostErrored
	}

	if status == 2 {
		return ErrHostShutdown
	}

	if status != 1 {
		return ErrHostNotRunning
	}

	shutdownUnsafe()

	return nil
}

func IsRunning() bool {
	mu.RLock()
	defer mu.RUnlock()

	return status == 1
}

func IsShutdown() bool {
	mu.RLock()
	defer mu.RUnlock()

	return status == 2 || status == 86
}

func AwaitShutdown() {
	wg.Wait()
}

func shutdownUnsafe() {
	var tmp []types.Service

	for _, s := range servicesById {
		var s2 = s()
		if s2 != nil {
			tmp = append(tmp, s2)
			v, ok := s2.(types.BackgroundService)
			if ok {
				v.Shutdown()
			}
			c, ok := s2.(cm.Closeable)
			if ok {
				c.Close()
			}
		}
	}

	servicesById = make(map[int32]types.ServiceFactory)
	servicesByName = make(map[string]types.ServiceFactory)

	for _, service := range tmp {
		v, ok := service.(types.BackgroundService)
		if ok {
			v.AwaitShutdown()
		}
		wg.Done()
	}

	status = 2
}

func doStartup(config cfg.Config) error {
	mu.Lock()
	defer mu.Unlock()

	if status == 86 {
		return ErrHostErrored
	}

	if status == 1 {
		return ErrHostAlreadyRunning
	}

	if status == 2 {
		return ErrHostShutdown
	}

	// Create each service in order to make available via GetService
	for _, s := range servicesById {
		var s2 = s()
		if s2 != nil {
			continue
		}

		shutdownUnsafe()
		status = 86

		return ErrServiceIsNil
	}

	// Confgure each service
	for n, s := range servicesByName {
		var s2 = s()
		if s2 != nil {
			v, ok := s2.(types.Configurable)
			if ok {
				v.Configure(config.Select(n))
			}
		}
	}

	status = 1

	return nil
}

func getServiceByNameUnsafe[T any](name string) (service T, e error) {
	if s, ok := servicesByName[name]; ok {
		service = s().(T)
		return
	}

	e = ErrServiceNotFound

	return
}

func getServiceByIdUnsafe[T any](id int32) (service T, e error) {
	if s, ok := servicesById[id]; ok {
		service = s().(T)
		return
	}

	e = ErrServiceNotFound

	return
}
