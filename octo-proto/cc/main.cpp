#include <iostream>
#include <fstream>
#include <string>
//#include <direct.h>
//#include <sys/stat.h>
#include "data/data.pb.h"
using namespace std;
//using namespace octo;

extern void mkdir_exec();

extern bool save_test();
extern bool load_test();

extern bool load();
extern bool change(int argc, char *argv[]);
extern bool save();

static octo::Data s_one_data;

int main(int argc, char *argv[])
{
    cout << "main - argc:" << to_string(argc) << endl;
    for (int i = 0; i < argc; i++) {
        cout << " - argv[" << to_string(i) << "]: = " << argv[i] << endl;
    }
    
    if(argc > 1)
    {
        string argv_1 = string(argv[1]);
        
        string str;
        if(argv_1=="load")
        {
            // load.
            cout << "-- load --" << endl;
            load_test();
        }
        else
        if(argv_1=="save")
        {
            // save.
            cout << "-- save --" << endl;
            save_test();
        }
        else
        if(argv_1=="change")
        {
            // change.
            cout << "-- change --" << endl;
            load();
            change(argc, argv);
        }
    }
}

void mkdir_exec()
{
/*
    struct stat buf;
    if(stat("file",&buf)==0)
    {
        cout << "-- folder exist. --" << endl;
        return;
    }
    mode_t mode = ;
    if(mkdir("./file", mode)==0){
        cout << "-- mkdir_exec success! --" << endl;
    }else{
        cout << "-- mkdir_exec failed. --" << endl;
    }
*/
    

/*
    string buffer;
    if(_stat("./file", &buffer)==0)
    {
        cout << "-- folder exist. --" << endl;
        return;
    }
    if(_mkdir("file")==0){
        cout << "-- mkdir_exec success! --" << endl;
    }else{
        cout << "-- mkdir_exec failed. --" << endl;
    }
*/
}

bool load()
{
    // ファイル読み込み
    cout << "--- load one. ---" << endl;
    std::fstream one_input("./file/octotest_onedata.bin", std::ios::in | std::ios::binary);
    if (!s_one_data.ParseFromIstream(&one_input)) {
        cout << "Failed to parse one data." << endl;
        return false;
    }
    cout << s_one_data.DebugString().c_str() << endl;
    
    // print.
    cout << "-- tag check -- (size:" << s_one_data.tag_size() << ")" << endl;
    
    for(int i=0; i<s_one_data.tag_size(); i++)
    {
        cout << "-- tag no[" + to_string(i) + "] - " << s_one_data.tag(i) << endl;
    }
    return true;
}

bool change(int argc, char *argv[])
{
    if(argc <= 2)
    {
        cout << "--- don't change. ---" << endl;
        return false;
    }
    
    cout << "--- change one. ---" << endl;
    
    string argv_name = string(argv[2]);
    
    // name change.
    s_one_data.set_name(argv_name);
    
    // tag change.
    s_one_data.set_tag(0, argv_name);
    
    save();
    
    return true;
}

bool save()
{
    mkdir_exec();
    
    cout << "--- save one. ---" << endl;

    //------
    cout << s_one_data.DebugString().c_str() << endl;
    
    int size = s_one_data.ByteSize();
    cout << " - data size : " << to_string(size) << endl;
    
    fstream one_output("./file/octotest_onedata.bin", ios::out | ios::trunc | ios::binary);
    if (!s_one_data.SerializeToOstream(&one_output)) {
        cerr << "Failed to write one data." << endl;
        return false;
    }
    
    return true;
}


bool load_test()
{
    // ファイル読み込み
    cout << "--- load test one. ---" << endl;
    octo::Data one_data;
    std::fstream one_input("./file/octotest_onedata.bin", std::ios::in | std::ios::binary);
    if (!one_data.ParseFromIstream(&one_input)) {
        cout << "Failed to parse one data." << endl;
        return false;
    }
    
    cout << one_data.DebugString().c_str() << endl;
    
    cout << "--- load test database. ---" << endl;
    octo::Database test_data;
    std::fstream db_input("./file/octotest_database.bin", std::ios::in | std::ios::binary);
    if (!test_data.ParseFromIstream(&db_input)) {
        cout << "Failed to parse db." << endl;
        return false;
    }
    cout << test_data.DebugString().c_str() << endl;
    
    return true;
}

bool save_test()
{
    mkdir_exec();
    
    cout << "--- save test. ---" << endl;
    
    //------
    octo::Data one_data;
    one_data.set_filepath("test_path");
    one_data.set_name("test_name");
    one_data.set_size(123);
    one_data.set_crc(4567);
    one_data.set_priority(0);
    one_data.add_tag("test_tag");
    one_data.add_tag("test_tag2");
    one_data.add_dependencie("test_dependencie");
    one_data.add_dependencie("test_dependencie2");
    one_data.set_state(octo::Data_State_NONE);
    
    cout << one_data.DebugString().c_str() << endl;
    
    int size = one_data.ByteSize();
    cout << " - data size : " << to_string(size) << endl;
    
    fstream one_output("./file/octotest_onedata.bin", ios::out | ios::trunc | ios::binary);
    if (!one_data.SerializeToOstream(&one_output)) {
        cerr << "Failed to write one data." << endl;
        return false;
    }
    
    //------
    
    octo::Database test_data;
    for(int i=0; i<3; i++)
    {
        test_data.set_revision(500 + i);
        
        octo::Data* oneP = test_data.add_list();
        string one_filepath = "filepath_" + to_string(i);
        string one_name = "name_" + to_string(i);
        
        oneP->set_filepath(one_filepath);
        oneP->set_name(one_name);
        oneP->set_size(200 + i);
        oneP->set_state(i%2 ? octo::Data_State_ADD : octo::Data_State_UPDATE);
    }
    
    cout << test_data.DebugString().c_str() << endl;
    
    fstream output("./file/octotest_database.bin", ios::out | ios::trunc | ios::binary);
    if (!test_data.SerializeToOstream(&output)) {
        cout << "Failed to write address book." << endl;
        return false;
    }
    
    return true;
}