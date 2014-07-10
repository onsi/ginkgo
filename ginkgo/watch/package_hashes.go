package watch

import "path/filepath"

type PackageHashes struct {
	PackageHashes map[string]*PackageHash
	usedPaths     map[string]bool
}

func NewPackageHashes() *PackageHashes {
	return &PackageHashes{
		PackageHashes: map[string]*PackageHash{},
		usedPaths:     nil,
	}
}

func (p *PackageHashes) CheckForChanges() []string {
	modified := []string{}

	for _, packageHash := range p.PackageHashes {
		if packageHash.CheckForChanges() {
			modified = append(modified, packageHash.path)
		}
	}

	return modified
}

func (p *PackageHashes) Add(path string) *PackageHash {
	path, _ = filepath.Abs(path)
	_, ok := p.PackageHashes[path]
	if !ok {
		p.PackageHashes[path] = NewPackageHash(path)
	}

	return p.Get(path)
}

func (p *PackageHashes) Get(path string) *PackageHash {
	path, _ = filepath.Abs(path)
	if p.usedPaths != nil {
		p.usedPaths[path] = true
	}
	return p.PackageHashes[path]
}

func (p *PackageHashes) StartTrackingUsage() {
	p.usedPaths = map[string]bool{}
}

func (p *PackageHashes) StopTrackingUsageAndPrune() {
	for path := range p.PackageHashes {
		if !p.usedPaths[path] {
			delete(p.PackageHashes, path)
		}
	}

	p.usedPaths = nil
}
