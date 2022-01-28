package base

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

var lockMap sync.Map

func CacheKey(path string, account *model.Account) string {
	path = utils.ParsePath(path)
	return fmt.Sprintf("%s%s", account.Name, path)
}

func SetCache(path string, obj interface{}, account *model.Account) error {
	cacheKey := CacheKey(path, account)
	err := conf.Cache.Set(conf.Ctx, cacheKey, obj, nil)
	v, ok := lockMap.Load(cacheKey)
	if ok {
		cond := v.(*sync.Cond)
		cond.Broadcast()
		lockMap.Delete(cacheKey)
	}
	return err
}

func GetCache(path string, account *model.Account) (interface{}, error) {
	cacheKey := CacheKey(path, account)
	value, err := conf.Cache.Get(conf.Ctx, cacheKey)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "value not found") {
		v, ok := lockMap.Load(cacheKey)
		if !ok {
			cond := sync.NewCond(&sync.Mutex{})
			lockMap.Store(cacheKey, cond)
			// 超时释放
			go time.AfterFunc(time.Second*5, func() {
				cond.Broadcast()
			})
			return nil, err
		} else {
			cond := v.(*sync.Cond)
			cond.Wait()
			return GetCache(cacheKey, account)
		}
	}
	return value, err
}

func DeleteCache(path string, account *model.Account) error {
	err := conf.Cache.Delete(conf.Ctx, CacheKey(path, account))
	log.Debugf("delete cache %s: %+v", path, err)
	return err
}
