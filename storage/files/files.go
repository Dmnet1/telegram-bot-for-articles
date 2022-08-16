package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"telegram-bot-for-articles/lib/e"
	"telegram-bot-for-articles/storage"
	"time"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774 // Права на чтение и запись.

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }() // Определяем способ обработки ошибок.

	fPath := filepath.Join(s.basePath, page.UserName) // Форомируем путь до директории, куда будет сохраняться файл.

	if err := os.MkdirAll(fPath, defaultPerm); err != nil { // Создаем все нужные директории.
		return err
	}

	fName, err := fileName(page) // Формируем имя файла.
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName) // Дописываем имя файла к пути.

	file, err := os.Create(fPath) // Создаем файл.
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil { // Записываем в файл страницу в нужном формате.
		return err
	}
	return nil
}

func (s Storage) PickRandom(UserName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't pick random page", err) }() // Определяем способ обработки ошибок.

	path := filepath.Join(s.basePath, UserName) // Форомируем путь до директории, куда будет сохраняться файл.

	files, err := os.ReadDir(path) // Получаем список файлов с помощью функции os.ReadDir.
	if err != nil {
		return nil, err
	}

	if len(files) == 0 { // Проверяет, что в директории со страницами есть сохраненные и в случае отсутствия, будет выведено сообщение.
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove file", err)
	}
	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file %s", path)

		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exist", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exist", path)

		return false, e.Wrap(msg, err)
	}

	return true, nil

}

func (s Storage) decodePage(filePath string) (*storage.Page, error) { // Декодирование файла и возвращение его содержимого.
	f, err := os.Open(filePath) // Открываем файл.
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() { _ = f.Close() }() // Закрываем файл.

	var p storage.Page // Создаем переменную в которую будет декодирован файл.

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
