package util

import (
	"fmt"
	"os"
	"path"
	"sync"
)

const (
	packagesNoData = "set:packages-nodata"
	distsNoMetaKey = "set:dists-meta-missing"

	distSet          = "set:dists"
	providerSet      = "set:providers"
	packageV1Set     = "set:packagesV1"
	packageV1SetHash = "set:packagesV1-Hash"
	packageV2Set     = "set:packagesV2"
	versionsSet      = "set:versions"

	distQueue      = "queue:dists"
	distQueueRetry = "queue:dists-Retry"
	providerQueue  = "queue:providers"
	packageP1Queue = "queue:packagesV1"
	packageV2Queue = "queue:packagesV2"

	lastUpdateTimeKey          = "key:last-update"
	packagistLastModifiedKey   = "key:packagist-last-modified"
	localStableComposerVersion = "key:local-stable-composer-version"
	v2LastUpdateTimeKey        = "key:v2-lastTimestamp"
	packagesJsonKey            = "key:packages.json"
)

var (
	// Wg Concurrency control
	Wg sync.WaitGroup
)

func Test(ctx *Context) (err error) {
	_, err = ctx.redis.Ping().Result()
	if err != nil {
		return
	}

	fmt.Println("[✓] test redis passed")

	err = ctx.github.Test()
	if err != nil {
		return
	}

	fmt.Println("[✓] test github passed")

	_, err = ctx.ossBucket.ListObjects()
	if err != nil {
		return
	}

	fmt.Println("[✓] test OSS passed")

	return
}

// Execute the main processing logic
func Execute() {

	if len(os.Args) != 2 {
		panic("must pass into `packagist.yml` configurations")
	}

	configPath := os.Args[1]

	if !path.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			panic("working directory is not existing")
		}
		configPath = path.Join(wd, configPath)
	}

	// Load config
	conf, err := LoadConfig(configPath)
	if err != nil {
		panic("load configuration error: " + err.Error())
	}

	fmt.Printf("load configurations successfully(%s)\n", configPath)

	// Init context
	ctx, err := NewContext(conf)
	if err != nil {
		panic("init context error: " + err.Error())
	}

	err = Test(ctx)
	if err != nil {
		panic("context test error: " + err.Error())
	}

	fmt.Println("test configurations successfully")

	if _, ok := os.LookupEnv("JUST_TEST"); ok {
		return
	}

	// Synchronize composer.phar
	go ctx.SyncComposerPhar("[SyncComposerPhar]")

	// Synchronize packages.json
	go ctx.SyncPackagesJsonFile("[PackagesJson]")

	// Synchronize Meta for V2
	go ctx.SyncV2("[SyncV2]")

	// Update status
	go ctx.SyncStatus("[Status]")

	Wg.Add(1)

	for i := 0; i < 12; i++ {
		go ctx.SyncProvider(fmt.Sprintf("[SyncProvider.%d]", i))
	}

	for i := 0; i < 10; i++ {
		go ctx.SyncPackagesV1(fmt.Sprintf("[SyncPackagesV1.%d]", i))
	}

	for i := 0; i < 10; i++ {
		go ctx.SyncPackagesV2(fmt.Sprintf("[SyncPackagesV2.%d]", i))
	}

	// Sync the dists
	for i := 0; i < 30; i++ {
		go ctx.SyncDists(fmt.Sprintf("[SyncDists.%d]", i))
	}

	// // Re-Sync the failed dists
	// for i := 0; i < 1; i++ {
	// 	go ctx.SyncDistsRetry(fmt.Sprintf("DistsRetry[%d]", i))
	// }
}
