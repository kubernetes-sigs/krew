function resolvegrep () {
  if [[ $(uname) == "Darwin" ]]; then
    if [[ $(whereis ggrep) -eq 0 ]]; then
      ggrep $@
    else
      echo "GNU grep not found - please install GNU coreutils from brew.sh (brew install coreutils)"
      exit 1
    fi
  else
    grep $@
  fi
}
