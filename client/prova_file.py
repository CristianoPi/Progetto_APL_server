from time import sleep
def create_file():
    with open('prova.txt', 'w') as file:
        sleep(5)
        file.write('Questo è un file di prova.')

if __name__ == "__main__":
    create_file()