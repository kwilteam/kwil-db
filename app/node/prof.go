package node

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

type profMode string

const (
	profModeDisabled profMode = ""
	profModeHTTP     profMode = "http"
	profModeCPU      profMode = "cpu"
	profModeMem      profMode = "mem"
	profModeBlock    profMode = "block"
	profModeMutex    profMode = "mutex"
)

func startProfilers(mode profMode, pprofFile string) (func(), error) {
	if pprofFile == "" {
		pprofFile = fmt.Sprintf("kwild-%s.pprof", mode)
	}

	switch mode {
	case profModeHTTP:
		// http pprof uses http.DefaultServeMux, so we register a redirect
		// handler with the root path on the default mux.
		http.Handle("/", http.RedirectHandler("/debug/pprof/", http.StatusSeeOther))
		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				fmt.Printf("http.ListenAndServe: %v\n", err)
			}
		}()
		return func() {}, nil
	case profModeCPU:
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, err
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return nil, fmt.Errorf("error starting CPU profiler: %w", err)
		}
		return pprof.StopCPUProfile, nil
	case profModeMem:
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, err
		}
		timer := time.NewTimer(time.Second * 15)
		go func() {
			<-timer.C
			if err = pprof.WriteHeapProfile(f); err != nil {
				fmt.Printf("WriteHeapProfile: %v\n", err)
			}
			f.Close()
		}()
		return func() { timer.Reset(0) }, nil
	case profModeBlock:
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, fmt.Errorf("could not create block profile file %q: %v", pprofFile, err)
		}
		runtime.SetBlockProfileRate(1)
		return func() {
			pprof.Lookup("block").WriteTo(f, 0)
			f.Close()
			runtime.SetBlockProfileRate(0)
		}, nil
	case profModeMutex:
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, fmt.Errorf("could not create mutex profile file %q: %v", pprofFile, err)
		}
		runtime.SetMutexProfileFraction(1)
		return func() {
			if mp := pprof.Lookup("mutex"); mp != nil {
				mp.WriteTo(f, 0)
			}
			f.Close()
			runtime.SetMutexProfileFraction(0)
		}, nil
	case profModeDisabled:
		return func() {}, nil
	default:
		return nil, fmt.Errorf("unknown profile mode %s", mode)
	}
}
