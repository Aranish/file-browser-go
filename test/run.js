/* Check if it works */
const
  child_process = require('child_process'),
  FileBrowser   = require('../lib/FileBrowser'),
  assert        = require('assert'),
  fs            = require('fs'),
  _             = require('lodash');

let browser, shell;

const TEST_HOME = process.env.TEST_HOME

describe ('Basic', function(){

  before(function(){
    shell = child_process.spawn ('./file-browser-go', [] , {
      stdio: ['pipe', 'pipe', 'ignore']
    });
    browser = new FileBrowser(shell);
  });

  it('can list contents of a directory', async function() {
    try{
      let result = await browser.ls(TEST_HOME + '/ls');
      assert(result.error.length === 0);
      assert(result.files.length === 4);
      return null;
    }catch(err){
      return err;
    }
  });

  it('can remove a directory', async function() {
    try{

      assert(fs.existsSync(TEST_HOME + '/rmdir'));
      assert(fs.existsSync(TEST_HOME + '/rmfile'));

      let result = await browser.rm(TEST_HOME + '/rmdir');
      assert(result.error.length === 0);
      
      result = await browser.rm(TEST_HOME + '/rmfile');
      assert(result.error.length === 0);

      assert(!fs.existsSync(TEST_HOME + '/rmdir'));
      assert(!fs.existsSync(TEST_HOME + '/rmfile'));
      return null;
      
    }catch(err){

      return err;

    }
  });

  it('can create a directory', async function() {
    try{

      assert(!fs.existsSync(TEST_HOME + '/mkdir'));
      let result = await browser.mkdir(TEST_HOME + '/mkdir');
      assert(result.error.length === 0);
      assert(fs.existsSync(TEST_HOME + '/mkdir'));
      return null;

    }catch(err){
      return err;
    }
  });

  it('can copy a directory', async function() {
    try{

      let result = await browser.cp(TEST_HOME + '/cpsrc', TEST_HOME + '/cpdest');
      assert(fs.existsSync(TEST_HOME + '/cpdest/cpsrc'));
      return null;

    }catch(err){

      return err;
    }
  });

  it('can move a directory', async function() {
    try{
      
      assert(fs.existsSync(TEST_HOME + '/mvdir'));
      assert(fs.existsSync(TEST_HOME + '/mvfile'));

      let result = await browser.mv(TEST_HOME + '/mvfile', TEST_HOME + '/mvdir/mvfile');
      assert(result.error.length === 0);

      assert(fs.existsSync(TEST_HOME + '/mvdir/mvfile'));

      result = await browser.mv(TEST_HOME + '/mvdir', TEST_HOME + '/mvdest');
      assert(result.error.length === 0);

      assert(fs.existsSync(TEST_HOME + '/mvdest'));

      return null;

    }catch(err){
      return err;
    }
  });

});