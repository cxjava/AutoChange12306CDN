#!/bin/bash
find ./dist -xdev -maxdepth 3 -type f -name 'AutoChange12306CDN*' | xargs upx