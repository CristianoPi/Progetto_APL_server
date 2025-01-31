def create_file():
    with open('prova5.txt', 'w') as file:
        file.write('Questo è un file di prova.')


if __name__ == "__main__":
    print('Ciao!')  
    create_file()