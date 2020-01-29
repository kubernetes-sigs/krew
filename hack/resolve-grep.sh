function resolvegrep () {
  if [[ $(uname) == "Darwin" ]]; then
    if [[ $(whereis ggrep) -eq 0 ]]; then
      ggrep $@
    else
      echo "GNU grep not found!"
      exit 1
    fi
  else
    grep $@
  fi
}
