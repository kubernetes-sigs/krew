function resolvelink () {
  if [[ $(uname) == "Darwin" ]]; then
    if [[ $(whereis greadlink) -eq 0 ]]; then
      greadlink -f $@
    else
      echo "GNU readlink not found - please install GNU coreutils from brew.sh (brew install coreutils)"
      exit 1
    fi
  else
    readlink -f $@
  fi
}
