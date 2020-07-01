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

	// Update status
	go status("Status", 1)

	Wg.Add(1)

	for i := 0; i < 13; i++ {
		go providers("Provider", i)
	}

	for i := 0; i < 60; i++ {
		go packagesP1("PackagesP1", i)
	}

	for i := 0; i < 60; i++ {
		go packagesP2("PackagesP2", i)
	}

	for i := 0; i < 60; i++ {
		go packagesP2Dev("PackagesP2Dev", i)
	}

	for i := 0; i < 60; i++ {
		go dists("Dists", i)
	}

	for i := 0; i < 1; i++ {
		go distsRetry("DistsRetry", i)
	}

}
