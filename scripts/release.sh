#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
RELEASE_SCRIPT_FILENAME="release-repo.sh"     # NOTE: Must be on the path; comes from devtools repo

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! bash "${RELEASE_SCRIPT_FILENAME}" "${root_dirpath}"; then
    echo "Error: Couldn't cut the release" >&2
    exit 1
fi
