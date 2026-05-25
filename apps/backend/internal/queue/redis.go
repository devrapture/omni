package queue

import (
	"crypto/tls"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/devrapture/omni/internal/config"
	"github.com/hibiken/asynq"
)

func RedisClientOpt(cfg *config.Config) asynq.RedisClientOpt {
	redisURL := cfg.REDIS_URL
	opt := asynq.RedisClientOpt{
		Addr: redisURL,
	}

	parsedURL, err := url.Parse(redisURL)
	if err != nil || parsedURL.Host == "" {
		return opt
	}

	opt.Addr = parsedURL.Host
	opt.DB = redisDB(parsedURL.Path)

	if parsedURL.User != nil {
		opt.Username = parsedURL.User.Username()
		opt.Password, _ = parsedURL.User.Password()
	}

	if parsedURL.Scheme == "rediss" {
		opt.TLSConfig = &tls.Config{
			ServerName: strings.Split(parsedURL.Host, ":")[0],
			MinVersion: tls.VersionTLS12,
		}
	}

	if _, _, err := net.SplitHostPort(opt.Addr); err != nil {
		opt.Addr = net.JoinHostPort(opt.Addr, "6379")
	}

	return opt
}

func redisDB(path string) int {
	path = strings.Trim(path, "/")
	if path == "" {
		return 0
	}

	db, err := strconv.Atoi(path)
	if err != nil {
		return 0
	}
	return db
}
