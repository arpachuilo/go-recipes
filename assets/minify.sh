#!/bin/sh
esbuild --bundle --minify --sourcemap ./css/*.css --outdir=./static/css/
esbuild --bundle --minify --sourcemap ./css/themes/*.css --outdir=./static/css/themes/
esbuild --bundle --minify --sourcemap ./js/*.js --outdir=./static/js/
