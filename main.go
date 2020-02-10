package main

import (
	"github.com/hnakamur/errstack"
	"github.com/hnakamur/ltsvlog/v3"
	"github.com/rs/xid"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var instances = make([]string, 0)
var cidrs = make([]*net.IPNet, 0)
var denyCode = 451
var nullBody = []byte("")

func main() {
	ltsvlog.Logger.Debug().String("event", "mastoguard start").Log()

	t := os.Getenv("PROXY_TARGET")
	if t == "" {
		ltsvlog.Logger.Err(errstack.WithLV(errstack.New("env 'PROXY_TARGET' must be set")))
		os.Exit(1)
	}
	ltsvlog.Logger.Debug().String("env 'PROXY_TARGET'", t).Log()

	h := os.Getenv("PROXY_HOST")
	ltsvlog.Logger.Debug().String("env 'PROXY_HOST'", h).Log()

	a := os.Getenv("LISTEN_ADDR")
	if a == "" {
		a  = ":8080"
	}
	ltsvlog.Logger.Debug().String("env 'LISTEN_ADDR'", a).Log()

	c := os.Getenv("DENY_CODE")
	if c != "" {
		if i, err := strconv.Atoi(c); err == nil && http.StatusText(i) != "" {
			denyCode = i
		} else {
			ltsvlog.Logger.Err(errstack.WithLV(errstack.New("env 'DENY_CODE' is invalid")))
			os.Exit(1)
		}
	}
	ltsvlog.Logger.Debug().String("env 'DENY_CODE'", c).Int("deny code", denyCode).Log()

	bs := os.Getenv("DENY_UA")
	if bs != "" {
		instances = strings.Split(bs, ",")
	} else {
		instances = make([]string, 0)
	}
	ltsvlog.Logger.Debug().String("env 'DENY_UA'", bs).Log()

	cs := os.Getenv("DENY_CIDR")
	if cs != "" {
		for _, v := range strings.Split(cs, ",") {
			_, ipnet, err := net.ParseCIDR(v)
			if err != nil {
				ltsvlog.Logger.Err(err)
				ltsvlog.Logger.Debug().String("cidr", v).Log()
			} else {
				cidrs = append(cidrs, ipnet)
			}
		}
	} else {
		cidrs = make([]*net.IPNet, 0)
	}
	ltsvlog.Logger.Debug().String("env 'DENY_CIDR'", cs).Log()

	u, err := url.Parse(t)
	if err != nil {
		ltsvlog.Logger.Err(err)
		return
	}

	p := httputil.NewSingleHostReverseProxy(u)
	http.HandleFunc("/", handler(p, u, h))
	ltsvlog.Logger.Info().String("event", "mastoguard ready").Log()
	ltsvlog.Logger.Err(http.ListenAndServe(a, nil))
}

func handler(p *httputil.ReverseProxy, u *url.URL, h string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		guid := xid.New().String()

		rip := remoteIP(r)
		sip := strings.Join(rip, ",")

		r.URL.Scheme = u.Scheme
		if h != "" {
			r.URL.Host = h
			r.Host = h
			r.Header.Set("Host", h)
		} else {
			r.URL.Host = u.Host
			r.Host = u.Host
			r.Header.Set("Host", u.Host)
		}
		f := r.Header.Get("X-Forwarded-For")
		if f == "" {
			f = r.RemoteAddr
		} else {
			f += "," + r.RemoteAddr
		}
		r.Header.Set("X-Forwarded-For", f)

		accessLog(r, guid, sip, "REQUEST")

		ua := r.UserAgent()
		for _, v := range instances {
			if strings.HasSuffix(ua, v) {
				accessLog(r, guid, sip, "DENY")
				defer accessLog(r, guid, sip, "HANDLED")

				w.WriteHeader(denyCode)
				_, _ = w.Write(nullBody)
				return
			}
		}

		for _, v := range cidrs {
			for _, u := range rip {
				if Contains(v, strings.TrimSpace(u)) {
					accessLog(r, guid, sip, "DENY")
					defer accessLog(r, guid, sip, "HANDLED")
					w.WriteHeader(denyCode)
					_, _ = w.Write(nullBody)
					return
				}
			}
		}

		w.Header().Set("Server", "mastoguard")

		accessLog(r, guid, sip, "ALLOW")
		defer accessLog(r, guid, sip, "HANDLED")
		p.ServeHTTP(w, r)
	}
}

func remoteIP(r *http.Request) []string {
	t := make([]string, 0)
	a, _, _ := net.SplitHostPort(r.RemoteAddr)
	f := r.Header.Get("X-Forwarded-For")
	if f != "" {
		t = strings.Split(f, ",")
	} else {
		t = append(t, a)
	}
	return t
}

func accessLog(r *http.Request, guid string, remoteAddr string, status string) {
	ltsvlog.Logger.Info().String("xid", guid).String("method", r.Method).String("url", r.URL.String()).String("remote", remoteAddr).String("useragent", r.UserAgent()).String("status", status).Log()
}

func Contains(cidr *net.IPNet, target string) bool {
	ip := net.ParseIP(target)
	return cidr.Contains(ip)
}
