# The ultimate truth or dare game!

## Description
This is a simple truth or dare game that can be played by a group of people.

## Features
- The game can be played by a group of people.

## Installation
1. Clone the repository:
    ```sh
    git clone https://github.com/2Friendly4You/TruthOrDare.git
    cd TruthOrDare
    ```

2. Create a [.env](http://_vscodecontentref_/1) file in the root directory with the following content:
    ```env
    MYSQL_USER=root
    MYSQL_ROOT_PASSWORD=example
    MYSQL_PASSWORD=example
    MYSQL_DATABASE=itemsdb
    MYSQL_HOST=db
    MYSQL_PORT=3306
    APP_PORT=8080
    ```

3. Start the server using Docker Compose:
    ```sh
    docker-compose up --build
    ```

If you want to delete the database and start fresh, you can use the following command:
```sh
docker-compose down -v
```

## Usage
1. Open your web browser and navigate to [http://localhost](http://_vscodecontentref_/2).
2. Add players and start the game.
3. Use the buttons to select "Truth" or "Dare" and follow the prompts.

## Documentation
1. Generate Swagger docs manually:
    ```sh
    make docs
    ```

## Contributing
- Fork the repository
- Create a new branch
- Make necessary changes and commit those changes
- Push the changes to GitHub
- Submit your changes for review

## License
None for now. Will be added later.