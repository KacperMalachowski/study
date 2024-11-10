#!/bin/bash

file="phonebook.txt"

function addEntry() {
  echo $'Podaj imie:'
  read name
  echo $'Podaj nazwisko:'
  read surname
  echo $'Podaj numer telefonu:'
  read phone

  echo "$name $surname $phone" >> $file

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

  entry="$name $surname $phone"
  if grep -qF "$entry" "$file"; then
    echo "$(grep -F "$entry" "$file")"
    echo "Usunac wpis? (t/n)"
    read answer

    if [ $answer == "t" ]; then
      grep -vF "$entry" "$file" > temp && mv temp "$file"
    fi
    echo "Usunieto"
  fi
}

function editEntry() {
  echo $'Podaj imie:'
  read name
  echo $'Podaj nazwisko:'
  read surname
  echo $'Podaj numer telefonu:'
  read phone

  entry="$name $surname $phone"
  if grep -qF "$entry" "$file"; then
    echo "$(grep -F "$entry" "$file")"

    echo "Edytować wpis? (t/n)"
    read answer

    if [ $answer == "t" ]; then
      echo "Podaj imie:"
      read newName
      echo "Podaj nazwisko:"
      read newSurname
      echo "Podaj numer telefonu:"
      read newPhone

      sed -i "s/$entry/$newName $newSurname $newPhone/g" "$file"
      echo "Zmieniono"
    fi
  fi
}

function findEntry() {
  echo "Wpisz szukana fraze:"
  read search

  echo "Wyniki wyszukiwania:"
  grep -F "$search" "$file"

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
        addEntry; break
        ;;
      "Usuń wpis")
        removeEntry; break
        ;;
      "Edytuj wpis")
        editEntry; break
        ;;
      "Wyszukaj wpis")
        findEntry; break
        ;;
      "Zakończ")
        echo "Zamykam program"
        exit
        ;;
      *) echo "Nieprawidłowa opcja $REPLY"; break;;
    esac
  done
done

