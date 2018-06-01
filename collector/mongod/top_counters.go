package mongod

import (
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	topTimeSecondsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "top_time_seconds_total",
		Help:      "The top command provides operation time, in seconds, for each database collection",
	}, []string{"type", "database", "collection"})
	topCountTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "top_count_total",
		Help:      "The top command provides operation count for each database collection",
	}, []string{"type", "database", "collection"})
)

// TopStatsMap is a map of top stats.
type TopStatsMap map[string]TopStats

// TopCounterStats represents top counter stats.
type TopCounterStats struct {
	Time  float64 `bson:"time"`
	Count float64 `bson:"count"`
}

// TopStats top collection stats
type TopStats struct {
	Total     TopCounterStats `bson:"total"`
	ReadLock  TopCounterStats `bson:"readLock"`
	WriteLock TopCounterStats `bson:"writeLock"`
	Queries   TopCounterStats `bson:"queries"`
	GetMore   TopCounterStats `bson:"getmore"`
	Insert    TopCounterStats `bson:"insert"`
	Update    TopCounterStats `bson:"update"`
	Remove    TopCounterStats `bson:"remove"`
	Commands  TopCounterStats `bson:"commands"`
}

// Export exports the data to prometheus.
func (topStats TopStatsMap) Export(ch chan<- prometheus.Metric) {
	for collectionNamespace, topStat := range topStats {
		namespace := strings.Split(collectionNamespace, ".")
		database := namespace[0]
		collection := strings.Join(namespace[1:], ".")

		topStatTypes := reflect.TypeOf(topStat)
		topStatValues := reflect.ValueOf(topStat)

		for i := 0; i < topStatValues.NumField(); i++ {
			metricType := topStatTypes.Field(i).Name

			opCount := topStatValues.Field(i).Field(1).Float()

			opTimeMicrosecond := topStatValues.Field(i).Field(0).Float()
			opTimeSecond := opTimeMicrosecond / 1e6

			topTimeSecondsTotal.WithLabelValues(metricType, database, collection).Set(opTimeSecond)
			topCountTotal.WithLabelValues(metricType, database, collection).Set(opCount)
		}
	}

	topTimeSecondsTotal.Collect(ch)
	topCountTotal.Collect(ch)
}

// Describe describes the metrics for prometheus.
func (topStats TopStatsMap) Describe(ch chan<- *prometheus.Desc) {
	topTimeSecondsTotal.Describe(ch)
	topCountTotal.Describe(ch)
}
