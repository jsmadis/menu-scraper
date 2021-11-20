# Menu scraper
`Menu scraper` is an CLI application that scrapes restaurants for it's daily menus and prints them in the terminal.

## Getting started

```bash
git clone https://github.com/jsmadis/menu-scraper.git
cd menu-scraper
go build
```

## Running the CLI app
To show help on how to run the app run the following command:

```bash
./menu-scraper -help
```

Basic commands to run the app are:

- prints daily menu for the whole week:
```bash
./menu-scraper
```

- prints daily menu only for today:
```bash
./menu-scraper --today
```

- prints restaurant based on their `tag`
```bash
./menu-scraper --tag HW
```

- prints only specified restaurants
```bash
./menu-scraper --name Cap --name Suzies
```

## Adding new restaurants

In order to add new restaurant you need to create new entry inside `config/restaurants.yml` with name, url and css selector.
The css selector must point to the menu in order to obtain it.

