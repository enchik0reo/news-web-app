package memcached

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"newsWebApp/app/authService/internal/models"
	"newsWebApp/app/authService/internal/storage"
)

type WarmUper interface {
	GetInfo(ctx context.Context) (*models.UsersInfo, error)
}

type Storage struct {
	db WarmUper

	timeout time.Duration
	c       *memcache.Client
	log     *slog.Logger
}

func New(wu WarmUper, host, port string, timeout time.Duration, slog *slog.Logger) (*Storage, error) {
	s := Storage{
		db:      wu,
		timeout: timeout,
		log:     slog,
	}

	mc := memcache.New(fmt.Sprintf("%s:%s", host, port))

	if err := mc.Ping(); err != nil {
		return nil, err
	}

	s.c = mc

	s.warmUp()

	return &s, nil
}

func (s *Storage) CheckEmail(email string) (bool, error) {
	const op = "storage.memcached.CheckEmail"

	item, err := s.c.Get(email)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false, nil
		}
		s.log.Error(op, "err", err.Error(), "email", email)
		return false, err
	}

	if item == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func (s *Storage) CheckUserName(userName string) (bool, error) {
	const op = "storage.memcached.CheckUserName"

	item, err := s.c.Get(userName)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false, nil
		}
		s.log.Error(op, "err", err.Error(), "user name", userName)
		return false, err
	}

	if item == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func (s *Storage) SaveUser(userName, email string) error {
	const op = "storage.memcached.SaveUser"

	if err := s.c.Set(&memcache.Item{Key: userName, Value: []byte{}}); err != nil {
		s.log.Error(op, "err", err.Error(), "user name", userName)
		return err
	}

	errors := make([]string, 2)

	if err := s.c.Set(&memcache.Item{Key: email, Value: []byte{}}); err != nil {
		if err := s.c.Delete(userName); err != nil {
			s.log.Error(op, "err", err.Error(), "user name", userName)
			errors[1] = err.Error()
		}
		s.log.Error(op, "err", err.Error(), "email", email)
		errors[0] = err.Error()
	}

	switch {
	case errors[0] != "" && errors[1] == "":
		return fmt.Errorf(errors[0])
	case errors[0] == "" && errors[1] != "":
		return fmt.Errorf(errors[1])
	case errors[0] != "" && errors[1] != "":
		return fmt.Errorf("err save: %s; err delete: %s", errors[0], errors[1])
	default:
		return nil
	}
}

func (s *Storage) DeleteUser(userName, email string) error {
	const op = "storage.memcached.DeleteUser"

	if err := s.c.Delete(userName); err != nil {
		s.log.Error(op, "err", err.Error(), "user name", userName)
		return err
	}

	errors := make([]string, 2)

	if err := s.c.Delete(email); err != nil {
		if err := s.c.Set(&memcache.Item{Key: userName, Value: []byte{}}); err != nil {
			s.log.Error(op, "err", err.Error(), "user name", userName)
			errors[1] = err.Error()
		}
		s.log.Error(op, "err", err.Error(), "email", email)
		errors[0] = err.Error()
	}

	switch {
	case errors[0] != "" && errors[1] == "":
		return fmt.Errorf(errors[0])
	case errors[0] == "" && errors[1] != "":
		return fmt.Errorf(errors[1])
	case errors[0] != "" && errors[1] != "":
		return fmt.Errorf("err delete: %s; err save: %s", errors[0], errors[1])
	default:
		return nil
	}
}

func (s *Storage) CloseConn() error {
	return s.c.Close()
}

func (s *Storage) warmUp() {
	const op = "storage.memcached.warmUp"

	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	usersInfo, err := s.db.GetInfo(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoUsers) {
			return
		}
		s.log.Debug("Can't get users info from db", "op", op, "err", err.Error())
	}

	if len(usersInfo.Names) == 0 || len(usersInfo.Emails) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for _, name := range usersInfo.Names {
			if err := s.c.Set(&memcache.Item{Key: name, Value: []byte{}}); err != nil {
				s.log.Error("Can't warm up name memcached", "op", op, "err", err.Error())
			}
		}
		wg.Done()
	}()

	for _, email := range usersInfo.Emails {
		if err := s.c.Set(&memcache.Item{Key: email, Value: []byte{}}); err != nil {
			s.log.Error("Can't warm up email memcached", "op", op, "err", err.Error())
		}
	}

	wg.Wait()
}
