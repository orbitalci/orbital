/*
	taskbuilder is a module for handling werker task builder events. it will build werker tasks by consuming messages off the `taskbuilder` queue and add them to the werker queue for building.

	this should eventually be for all werker builder events, but for right now it will be specifically for events that are signaled by a subscriptions, ie a build that
	subscribes to an upstream repo. if the upstream repo builds successfully, events will be sent here for werker task generation on the downstream repo.
*/
package taskbuilder
