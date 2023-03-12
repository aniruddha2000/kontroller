#!/usr/bin/python3

versionFile = open("VERSION", "r+")

minor = versionFile.readline().split(".")
minor[2] = str(int(minor[2]) + 1)
newVersion = ".".join(minor)
versionFile.seek(0)
versionFile.truncate(0)

versionFile.write(newVersion)
versionFile.close()
