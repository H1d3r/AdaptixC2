#ifndef ADAPTIXCLIENT_SETTINGS_H
#define ADAPTIXCLIENT_SETTINGS_H

#include <main.h>

class DialogSettings;
class MainAdaptix;

class Settings
{
    MainAdaptix* mainAdaptix = nullptr;

public:
    explicit Settings(MainAdaptix* m);
    ~Settings();

    DialogSettings* dialogSettings = nullptr;
    SettingsData    data;

    MainAdaptix* getMainAdaptix();
    void SetDefault();
    void LoadFromDB();
    void SaveToDB() const;
};

#endif
