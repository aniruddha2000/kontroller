#!/usr/bin/python3

import argparse

versionFile = open("VERSION", "r+")
minor = versionFile.readline().split(".")

def increment_patch() -> str:
    minor[2] = str(int(minor[2]) + 1)
    return ".".join(minor)

def increment_minor() -> str:
    minor[1] = str(int(minor[1]) + 1)
    minor[2] = "0"
    return ".".join(minor)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--increment-minor", dest="incminor", default=False, action="store_true", help="Increment the Minor version in the VERSION file. e.g. 1.2.3 => 1.3.0")
    args = parser.parse_args()

    if args.incminor:
        newVersion = increment_minor()
    else:
        newVersion = increment_patch()

    versionFile.seek(0)
    versionFile.truncate(0)
    versionFile.write(newVersion)
    versionFile.close()
