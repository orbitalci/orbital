# This is to ensure `make` will run the container builds
.PHONY: admin hookhandler worker

# TODO: Should I spend the time to make this a generic make target?
admin:
	docker build -t ocelot-${@} -f ${@}/Dockerfile .

hookhandler:
	docker build -t ocelot-${@} -f ${@}/Dockerfile .

worker:
	docker build -t ocelot-${@} -f ${@}/Dockerfile .
