/*
valet is responsible for keeping track of all the active builds on the werker
There are two different types; the context valet and the (build) valet. The context valet manages active contexts of builds,
which are responsible for killing builds. The (build) valet handles build runtime data, i.e. registration with consul for new builds, changes in build stages,
and eventual cleanup of builds after they completed. This valet is also responsible for proper panic handling and storage.
*/
package valet
