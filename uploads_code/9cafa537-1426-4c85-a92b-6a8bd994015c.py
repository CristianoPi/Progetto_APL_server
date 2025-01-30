def create_file():
    with open('prova.txt', 'w') as file:
        file.write('Questo è un file di prova.')

if __name__ == "__main__":
    create_file()