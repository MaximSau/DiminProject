package app

import "pulsefit/internal/service"

type Catalog = service.Catalog

func LoadCatalog() (*Catalog, error) {
	return service.LoadCatalog()
}
