function resolvelink () {
  if [[ $(uname) == "Darwin" ]]; then
    if [[ $(whereis greadlink) -eq 0 ]]; then
      greadlink -f $@
    else
      echo "GNU readlink not found"
      exit 1
    fi
  else
    readlink -f $@
  fi
}
