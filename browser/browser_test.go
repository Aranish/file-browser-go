package browser;

import (
	"testing";
	"os";
	"path/filepath";
	"io";
	"io/ioutil";
	"math/rand";
	"gopkg.in/vmihailenco/msgpack.v2";
)

func TestList(t *testing.T) {
	temp, err := ioutil.TempDir("","list");
	FailNotNil(err, t);
	defer os.Remove(temp);
	// Make two directories in temp
	err = os.Mkdir(filepath.Join(temp, "a"), 0777);
	FailNotNil(err, t);
	err = os.Mkdir(filepath.Join(temp, "b"), 0777);
	FailNotNil(err, t);
	res := List("test", temp).(*ResultSet);
	if res.Err != "" {
		t.Fail();
	}
	if len(res.Files) != 2 {
		t.Fail();
	}
}

func TestListNotExist(t *testing.T) {
	res := List("test","does/not/exist").(*ResultSet);
	if res.Err == "" {
		t.Log("Error should occur.");
		t.Fail();
	}
	t.Log("Error: ", res.Err);
}

func TestMakeDirectoryAndRemove(t *testing.T) {
	home , err:= ioutil.TempDir("","MakeAndRemove");
	FailNotNil(err, t);
	paths := []string{"/test_folder","/test_folder/sub_folder"};

	defer func() {
		r := Remove("test", filepath.Join(home, paths[0])).(*ResultSet);
		if r.Err != "" {
			t.Log(r.Err);
			t.Fail();
		}
	}();

	for _, p := range paths {
		res := MakeDirectory("test", filepath.Join(home, p)).(*ResultSet);
		if res.Err != "" {
			t.Log(res.Err);
			t.Fail();
		}
		if !Exists(filepath.Join(home, p)) {
			t.Log("Directory not created");
			t.Fail();
		}
	}
}

func TestMakeDirectoryBadPath(t *testing.T) {
	res := MakeDirectory("test","does/not/exist").(*ResultSet);
	if res.Err == "" {
		t.Log("Error should occur.");
		t.FailNow();
	}
	t.Log("Error: ", res.Err);
}

func TestCopy (t *testing.T) {

	home, err := ioutil.TempDir("","copy");
	FailNotNil(err, t);

	dir := []string{"copy_folder", "copy_folder/sub1",
	"copy_folder/sub2", "copy_folder/sub1/sub3",
	"copy_to"};

	// Create a directory for copying
	for _, p := range dir {
		res := MakeDirectory("test", filepath.Join(home,p));
		if err := res.(*ResultSet).Err; err != "" {
			t.Log(err);
			t.FailNow();
		}
	}

	d1 := filepath.Join(home,"copy_folder");
	d2 := filepath.Join(home,"copy_to/");

	defer func() {
		_ = os.Remove(d1);
		_ = os.Remove(d2);
	}();

	_ = Copy("test", d1, d2, ioutil.Discard);
	WaitForOperationsToComplete();

	if CompareDirectory(d1, filepath.Join(d2, "copy_folder")) == false {
		t.Logf("Directories not similar.");
		t.Fail();
	}
}

func TestCopySrcNotExist (t *testing.T){
	res := Copy("test" ,"does/not/exist", os.TempDir(), ioutil.Discard).(*ResultSet);
	if res.Err == "" {
		t.Log("Error should occur.");
		t.FailNow();
	}
	t.Log("Error: ", res.Err);
}

func TestCopyDestNotExist(t *testing.T) {
	f, err := ioutil.TempFile("", "existent_source");
	FailNotNil(err, t);
	f.Close();
	res := Copy("test", f.Name(), "does/not/exist", ioutil.Discard).(*ResultSet);
	if res.Err == "" {
		t.Log("Error should occur.");
		t.FailNow();
	}
	t.Log("Error: ", res.Err);
}

func TestGetFile (t *testing.T) {

	var temp string;
	var tf, outputFile *os.File;

	defer func() {
		_ = os.Remove(temp);
		if tf != nil {
			_ = tf.Close();
		}
	}();

	temp, err := ioutil.TempDir("", "GetFile");
	FailNotNil(err, t);
	tf, err = ioutil.TempFile(temp, "getfile");
	FailNotNil(err, t);
	outputFile, err = os.OpenFile(filepath.Join(temp, "catch"), os.O_CREATE | os.O_WRONLY, 0777);
	FailNotNil(err, t);

	data := []byte{};

	gen := rand.New(rand.NewSource(1));
	size := 3000;
	for i := 0; i < int(size); i++ {
		num := gen.Int31();

		b := []byte{0,0,0,0};
		var k int64 = 3;
		for k >= 0 {
			b[k] = byte(num & 0xff);
			k--;
			num = num >> 8;
		}

		for _, j := range b {
			data = append(data, j);
		}
	}

	// File paths
	fp := tf.Name();
	op := outputFile.Name();

	defer func() {
		_ = os.Remove(fp);
		_ = os.Remove(op);
	}();

	// Write data to temp file and close
	_, err = tf.Write(data);
	FailNotNil(err, t);
	_ = tf.Close();

	// Get the file and write output to outputFile
	GetFile("test", fp, outputFile);
	WaitForOperationsToComplete();
	// Close the file after writing
	outputFile.Close();

	// Reopen file in READ ONLY mode
	outputFile, err = os.OpenFile(op, os.O_RDONLY, 0777);
	FailNotNil(err, t);

	// Create decoder from GetFile output
	msgpackDec := msgpack.NewDecoder(outputFile);
	err = nil;
	res := &ResultSet{};

	// var max int64 = -1;

	// Buffer to hold the data that's been read
	compBuff := []byte{};

	for err != io.EOF {
		err = msgpackDec.Decode(res);
		if err != nil {
			t.Log(err);
			break;
		}
		t.Log(res);
		if res.Data.CurrentPiece == res.Data.TotalPieces {
			t.Logf("Got total pieces %d", res.Data.TotalPieces);
		}
		for _, b := range res.Data.Data {
			compBuff = append(compBuff, b);
		}
	}

	outputFile.Close();

	for i, _ := range data {
		if data[i] != compBuff[i]{
			t.Logf("Data not same");
			t.FailNow();
		}
	}
}

func TestGetFileEmpty(t *testing.T) {
	var tf, of *os.File;
	var err error;
	tf, err = ioutil.TempFile("","getfile");
	FailNotNil(err, t);
	of, err = ioutil.TempFile("", "catch");

	_,_ = tf.Write([]byte{});
	_ = tf.Close();

	tp := tf.Name();
	op := of.Name();
	GetFile("test", tp, of);
	WaitForOperationsToComplete();

	of.Close();

	of, err = os.OpenFile(op, os.O_RDONLY, 0777);
	msgpackDec := msgpack.NewDecoder(of);

	// var max int64 = -1;
	err = nil;
	res := &ResultSet{};

	for err != io.EOF {
		err = msgpackDec.Decode(res);
		if err != nil {
			t.Log(err);
			break;
		}
		t.Log(res);
		if res.Err != "EOF" {
			t.FailNow();
		}
	}
}

func TestGetFileNotExist(t *testing.T) {
	path := "/this/is/not/a/valid/path";
	op, err := ioutil.TempFile("", "get_file_not_exist_op");
	FailNotNil(err, t);

	GetFile("test", path, op);
	WaitForOperationsToComplete();
	op.Close();

	op, err = os.Open(op.Name());
	FailNotNil(err, t);

	dec := msgpack.NewDecoder(op);
	var res *ResultSet = nil;
	dec.Decode(&res);
	if res.Err == "" {
		t.Log("Should return error.");
		t.FailNow();
	}

	t.Log("Error: ", res.Err);
}

func TestPutFile (t *testing.T) {
	data := []byte{};

	gen := rand.New(rand.NewSource(1));
	t.Log("Generating bytes: ");
	size := 3000;
	for i := 0; i < int(size); i++ {
		num := gen.Int31();

		b := []byte{0,0,0,0};
		var k int64 = 3;
		for k >= 0 {
			b[k] = byte(num & 0xff);
			k--;
			num = num >> 8;
		}

		for _, j := range b {
			data = append(data, j);
		}
	}
	t.Log("Bytes generated: ");

	f, err := ioutil.TempFile("", "put_file_test");
	FailNotNil(err, t);
	newpath := f.Name();
	os.Remove(newpath);
	t.Log(newpath);

	// Write data to newpath using PutFile
	var count int = 0;
	for count < len(data) {
		w := []byte{};
		i := 0;
		for i < CHUNKSIZE && count < len(data) {
			w = append(w, data[count]);
			i++;
			count++;
		}
		t.Log("Chunk Length: ", len(w));
		res := PutFile("test", newpath, w).(*ResultSet);
		if res.Err != "" {
			t.FailNow();

			_ = PutFile("test", newpath, []byte{});
		}
	}
	_ = PutFile("test", newpath, []byte{});
	file, err := os.OpenFile(newpath, os.O_RDONLY, 0777);
	defer os.Remove(newpath);
	FailNotNil(err, t);
	defer file.Close();
	dataCopy, err := ioutil.ReadAll(file);
	FailNotNil(err ,t);
	// t.Log(dataCopy);
	t.Log(len(dataCopy) , len(data));

	for i, _ := range data {
		if data[i] != dataCopy[i] {
			t.FailNow();
		}
	}
}

func TestPutFileEmpty(t *testing.T) {
	newpath := filepath.Join(os.TempDir(), "put_file_empty_test");
	PutFile("test", newpath, []byte{});
	PutFile("test", newpath, []byte{});
	file, err := os.Open(newpath);
	FailNotNil(err, t);
	defer file.Close();
	data, err := ioutil.ReadAll(file);
	FailNotNil(err, t);
	if len(data) != 0 {
		t.FailNow();
	}
}

func TestPutFileNotExists(t *testing.T) {
	path := "this/path/does/not/exist";
	res := PutFile("test", path, []byte{}).(*ResultSet);
	if res.Err == "" {
		t.Log("Error should occur");
		t.FailNow();
	}
	t.Log("Error: ", res.Err);
}
