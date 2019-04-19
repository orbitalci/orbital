/*
buildmonitor is responsible for keeping track of all the active builds on the werker
There are two different types; the BuildReaper and the BuildMonitor. The BuildReaper manages active contexts of builds,
which are responsible for killing builds. The BuildMonitor handles build runtime data, i.e. registration with consul for new builds, changes in build stages,
and eventual cleanup of builds after they completed. buildmonitor is also responsible for proper panic handling and storage.
*/
package buildmonitor
