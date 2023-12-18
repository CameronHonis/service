package service

// PROBLEMS TO SOLVE FOR:
// 1. Make it clear how services get updated via updates on config
// 2. Make it easy to update existing config

// SOLUTIONS:
// 1. Configs should be treated as immutable objects, and when it's time to update a services config, a new config object
//		should be explicitly injected into the service.
// 2. Configs should support merging, where the merging config will merge "on top" of the existing config. A default
//		method should NOT belong on ConfigI for two reasons:
//			1. Merging configs is usually not trivial, since there are no optional fields (only zero-values) in go
//			2. A base Config struct method will not have access to any useful fields when implemented on a real config
//				object.

type ConfigI interface {
	MergeWith(config ConfigI) ConfigI
}
