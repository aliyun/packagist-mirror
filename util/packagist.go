package util

type Packagist struct {
	repoUrl string
	apiUrl  string
}

func NewPackagist(repoUrl string, apiUrl string) (packagist *Packagist) {
	return &Packagist{
		repoUrl: repoUrl,
		apiUrl:  apiUrl,
	}
}

func (packagist *Packagist) GetPackagesJSON() (content []byte, err error) {
	url := packagist.repoUrl + "packages.json"
	content, err = GetBody(url)
	return
}

func (packagist *Packagist) GetMetadataChanges(lastTimestamp string) (content []byte, err error) {
	url := packagist.apiUrl + "metadata/changes.json?since=" + lastTimestamp
	content, err = GetBody(url)
	return
}

func (packagist *Packagist) GetInitMetadataChanges() (content []byte, err error) {
	url := packagist.apiUrl + "metadata/changes.json"
	content, err = GetBody(url)
	return
}

func (packagist *Packagist) GetAllPackages() (content []byte, err error) {
	url := packagist.apiUrl + "packages/list.json"
	content, err = GetBody(url)
	return
}

func (packagist *Packagist) GetPackage(packageName string) (content []byte, err error) {
	url := packagist.apiUrl + "p2/" + packageName + ".json"
	content, err = GetBody(url)
	return
}
