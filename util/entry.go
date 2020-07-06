package util

// Execute the main processing logic
func Execute() {

	// Load config
	loadConfig()

	// Init Redis Client
	initRedisClient()

	// Synchronize composer.phar
	go composerPhar("ComposerPhar", 1)

	// Synchronize packages.json
	go packagesJsonFile("PackagesJson")

	// Synchronize Meta for V2
	go syncV2("SyncV2")

	// Update status
	go status("Status", 1)

	Wg.Add(1)

	for i := 0; i < 13; i++ {
		go providers("Provider", i)
	}

	for i := 0; i < 60; i++ {
		go packagesV1("PackagesV1", i)
	}

	for i := 0; i < 60; i++ {
		go packagesV2("SyncV2PackagesV2", i)
	}

	for i := 0; i < 60; i++ {
		go dists("Dists", i)
	}

	for i := 0; i < 1; i++ {
		go distsRetry("DistsRetry", i)
	}

}
