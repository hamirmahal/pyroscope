package phlare

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	phlarecontext "github.com/grafana/pyroscope/pkg/phlare/context"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/services"
	"google.golang.org/grpc/health/grpc_health_v1"

	compactionworker "github.com/grafana/pyroscope/pkg/experiment/compactor"
	segmentwriter "github.com/grafana/pyroscope/pkg/experiment/ingester"
	"github.com/grafana/pyroscope/pkg/experiment/metastore"
	metastoreclient "github.com/grafana/pyroscope/pkg/experiment/metastore/client"
	"github.com/grafana/pyroscope/pkg/experiment/querybackend"
	querybackendclient "github.com/grafana/pyroscope/pkg/experiment/querybackend/client"
	"github.com/grafana/pyroscope/pkg/util/health"
)

func (f *Phlare) initSegmentWriter() (services.Service, error) {
	phlarectx := f.context()
	reg := phlarecontext.Registry(phlarectx)
	reg = prometheus.WrapRegistererWith(prometheus.Labels{"component": "segment-writer"}, reg)
	log := phlarecontext.Logger(phlarectx)
	ingester, err := segmentwriter.New(
		reg,
		log,
		f.Cfg.SegmentWriter,
		f.Cfg.PhlareDB,
		f.storageBucket, f.metastoreClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create segment writer: %w", err)
	}
	f.segmentWriter = ingester
	f.API.RegisterSegmentWriter(ingester)
	return f.segmentWriter, nil
}

func (f *Phlare) initCompactionWorker() (svc services.Service, err error) {
	logger := log.With(f.logger, "component", "compaction-worker")
	f.compactionWorker, err = compactionworker.New(f.Cfg.CompactionWorker, logger, f.metastoreClient, f.storageBucket, f.reg)
	if err != nil {
		return nil, err
	}
	return f.compactionWorker, nil
}

func (f *Phlare) initMetastore() (services.Service, error) {
	logger := log.With(f.logger, "component", "metastore")
	m, err := metastore.New(f.Cfg.Metastore, f.TenantLimits, logger, f.reg, f.healthService, f.metastoreClient)
	if err != nil {
		return nil, err
	}
	f.API.RegisterMetastore(m)
	f.metastore = m
	return m.Service(), nil
}

func (f *Phlare) initMetastoreClient() (services.Service, error) {
	mc, err := metastoreclient.New(f.Cfg.Metastore.Address, f.Cfg.Metastore.GRPCClientConfig, f.logger)
	if err != nil {
		return nil, err
	}
	f.metastoreClient = mc
	return mc.Service(), nil
}

func (f *Phlare) initQueryBackend() (services.Service, error) {
	br := querybackend.NewBlockReader(f.logger, f.storageBucket)
	logger := log.With(f.logger, "component", "query-backend")
	b, err := querybackend.New(f.Cfg.QueryBackend, logger, f.reg, f.queryBackendClient, br)
	if err != nil {
		return nil, err
	}
	f.API.RegisterQueryBackend(b)
	return b.Service(), nil
}

func (f *Phlare) initQueryBackendClient() (services.Service, error) {
	c, err := querybackendclient.New(
		f.Cfg.QueryBackend.Address,
		f.Cfg.QueryBackend.GRPCClientConfig,
	)
	if err != nil {
		return nil, err
	}
	f.queryBackendClient = c
	return c.Service(), nil
}

func (f *Phlare) initHealthService() (services.Service, error) {
	healthService := health.NewGRPCHealthService()
	grpc_health_v1.RegisterHealthServer(f.Server.GRPC, healthService)
	f.healthService = healthService
	return healthService, nil
}