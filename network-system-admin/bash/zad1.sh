#!/bin/bash

entries=()

function addEntry() {
  echo $'Podaj imie:'
  read name
  echo $'Podaj nazwisko:'
  read surname
  echo $'Podaj numer telefonu:'
  read phone

  entries+=("$name $surname $phone")

  echo "Dodać kolejny wpis? (t/n)"
  read answer
  if [ $answer == "t" ]; then
    addEntry
  fi
}

function removeEntry() {
  echo $'Podaj imie:'
  read name
  echo $'Podaj nazwisko:'
  read surname
  echo $'Podaj numer telefonu:'
  read phone

  for i in "${!entries[@]}"; do
    if [[ "${entries[$i]}" == "$name $surname $phone" ]]; then
      echo "${entries[$i]}"
      echo "Usunac wpis? (t/n)"
      read answer

      if [ $answer == "t" ]; then
        unset entries[$i]
        echo "Usunieto"
      fi
      break
    fi
  done
}

function editEntry() {
  echo $'Podaj imie:'
  read name
  echo $'Podaj nazwisko:'
  read surname
  echo $'Podaj numer telefonu:'
  read phone

  for i in "${!entries[@]}"; do
    if [[ "${entries[$i]}" == "$name $surname $phone" ]]; then
      echo "${entries[$i]}"

      echo "Edytować wpis? (t/n)"
      read answer

      if [ $answer == "n" ]; then
        break
      fi

      echo "Podaj imie:"
      read newName
      echo "Podaj nazwisko:"
      read newSurname
      echo "Podaj numer telefonu:"
      read newPhone

      entries[$i]="$newName $newSurname $newPhone"
      echo "Zmieniono"
      break
    fi
  done
}

function findEntry() {
  echo "Wpisz szukana fraze:"
  read search

  echo "Wyniki wyszukiwania:"
  for i in "${!entries[@]}"; do
    if [[ "${entries[$i]}" == *"$search"* ]]; then
      echo "${entries[$i]}"
    fi
  done

  echo "Wyszukać ponownie? (t/n)"
  read answer

  if [ $answer == "t" ]; then
    findEntry
  fi
}

while true; do
  echo "Menu główne:"
  options=("Dodaj wpis" "Usuń wpis" "Edytuj wpis" "Wyszukaj wpis" "Zakończ")
  PS3=$'Co chcesz zrobić? \n'
  select opt in "${options[@]}"
  do
    case $opt in
      "Dodaj wpis")
        addEntry
        ;;
      "Usuń wpis")
        removeEntry
        ;;
      "Edytuj wpis")
        editEntry
        ;;
      "Wyszukaj wpis")
        findEntry
        ;;
      "Zakończ")
        echo "Zamykam program"
        exit
        ;;
      *) echo "Nieprawidłowa opcja $REPLY";;
    esac
  done
done

