# Yummy

![RecipeMaster Logo](https://via.placeholder.com/150)

**yummy** is a powerful command-line interface (CLI) tool designed to help you manage your recipes effortlessly. With yummy, you can store, view, categorize, and export your recipes using a beautiful interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Store Recipes**: Easily add new recipes with details like name, ingredients, instructions, and cooking time.
- **View Recipes**: Browse through your collection of recipes or search for specific ones.
- **Categorize Recipes**: Organize your recipes into categories for better management.
- **Export Recipes**: Export your recipes to various formats like JSON or CSV.
- **Beautiful Interface**: Enjoy a user-friendly interface thanks to Bubble Tea and other dependencies.

## Installation

To install yummy, you need to have Go installed on your machine. Follow these steps:

1. **Clone the Repository**:
   ```sh
   git clone https://github.com/yourusername/yummy.git
   cd yummy
   ```

2. **Build the Project**:
   ```sh
   go build -o yummy
   ```

3. **Run the CLI Tool**:
   ```sh
   ./yummy
   ```

## Usage

yummy provides a simple and intuitive command-line interface. Here are some examples of how to use it:

- **Add a New Recipe**:
  ```sh
  ./yummy add --name \"Pancakes\" --ingredients \"Flour, Milk, Eggs\" --instructions \"Mix and cook\" --time 15
  ```

- **View All Recipes**:
  ```sh
  ./yummy view
  ```

- **Categorize a Recipe**:
  ```sh
  ./yummy categorize --name \"Pancakes\" --category \"Breakfast\"
  ```

- **Export Recipes**:
  ```sh
  ./yummy export --format json
  ```

## Dependencies

RecipeMe leverages the following dependencies to provide a seamless experience:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea): For building the beautiful CLI interface.

## Contributing

Contributions are welcome! If you have any ideas, suggestions, or bug reports, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

If you have any questions or need further assistance, feel free to reach out to us at [ad.pesek13@gmail.com](mailto:ad.pesek13@gmail.com).

---

**Happy cooking with yummy!** 🍳🍴
