PWD := ${CURDIR}
DOCKER_IMAGE_NAME := "wwwwildcat"
CONTAINER_REMOVE_FLAG := "--rm"
BASE_URL := "https://tamada.github.io/wildcat"
HUGO_THEME := "cayman-hugo-theme"
JOJOMI_VERSION := 0.69.0


build: themes/cayman-hugo-theme/README.md public
	docker run ${CONTAINER_REMOVE_FLAG} --name ${DOCKER_IMAGE_NAME}_build -v "${PWD}":/src -v ${PWD}/public:/output -e HUGO_THEME=$(HUGO_THEME) -e HUGO_BASEURL=${BASE_URL} jojomi/hugo:${JOJOMI_VERSION}
	rm public/favicon* public/apple-touch-icon.png

public:
	git worktree prune
	git worktree add public gh-pages

themes/cayman-hugo-theme/README.md:
	git submodule update --init

start: themes/cayman-hugo-theme/README.md
	docker run ${CONTAINER_REMOVE_FLAG} -d --name ${DOCKER_IMAGE_NAME} -p 1313:1313 -v "${PWD}":/src -v "$(PWD)"/public:/output -e HUGO_THEME=$(HUGO_THEME) -e HUGO_WATCH="true" -e HUGO_BASEURL=${BASE_URL} jojomi/hugo:${JOJOMI_VERSION}

stop:
	docker stop ${DOCKER_IMAGE_NAME}
