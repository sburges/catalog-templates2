#!/bin/sh -x

OS_VERSION=$(go env GOOS)


if [ -z ${APP_VERSION_TAG} ]; then
	echo "Error: APP_VERSION_TAG is not set"
	exit 1
fi

if [ -z ${VERSION} ]; then
	echo "Error: VERSION is not set"
	exit 1
fi

SERVICE_NAME=$1

if [ -z ${SERVICE_NAME} ]; then
	echo "Error: Please secify the service name to be updated"
	exit 1
fi

REGISTRY_LOCATION=docker-registry.helion.space:443
IMAGE_TAG=latest
SERVICE_VERSION=latest

for file_path in `grep -l ${REGISTRY_LOCATION} services/HPE/${SERVICE_NAME} -r`
do
	grep ${REGISTRY_LOCATION} ${file_path} | grep csm
	if [ $? -eq 0 ]; then
		#darwin
		if [ "$OS_VERSION" = "darwin" ]; then
			sed -i '.original' "\|${REGISTRY_LOCATION}/catalog-service-manager/csm|s|${IMAGE_TAG}|${APP_VERSION_TAG}|gwp" ${file_path}
			sed -i '.updated_image' "\|version|s|${SERVICE_VERSION}|${VERSION}|gwp" ${file_path}
		elif [ "$OS_VERSION" = "linux" ]; then
			sed -i "\|${REGISTRY_LOCATION}/catalog-service-manager/csm|s|${IMAGE_TAG}|${APP_VERSION_TAG}|g" ${file_path}
			sed -i "\|version|s|${SERVICE_VERSION}|${VERSION}|g" ${file_path}
		fi

	fi
done
