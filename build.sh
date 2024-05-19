#!/bin/bash

set -e

lint-md ./*.md -f
lint-md ./*/*.md -f
