#include <UI/Dialogs/DialogSettings.h>

DialogSettings::DialogSettings(Settings* s)
{
    settings = s;

    this->createUI();

    connect(listSettings, &QListWidget::currentRowChanged, this, &DialogSettings::onStackChange);
    connect(buttonApply, &QPushButton::clicked, this, &DialogSettings::onApply);

}

void DialogSettings::createUI()
{
    this->setWindowTitle("Adaptix Settings");
    this->resize(1200, 600);

    /////////////// Main setting

    QWidget*     mainSettingWidget = new QWidget(this);
    QGridLayout* mainSettingLayout = new QGridLayout(mainSettingWidget);

    QLabel* themeLabel = new QLabel(mainSettingWidget);
    themeLabel->setText("Main theme: ");

    QComboBox*   themeCombo = new QComboBox(mainSettingWidget);
    themeCombo->addItem("Light_Arc");
    themeCombo->addItem("Dark");

    QCheckBox* consoleTimeCheck = new QCheckBox(mainSettingWidget);
    consoleTimeCheck->setText("Print date and time in agent console");
    consoleTimeCheck->setChecked(true);

    mainSettingLayout->addWidget(themeLabel, 0, 0, 1, 1);
    mainSettingLayout->addWidget(themeCombo, 0, 1, 1, 1);
    mainSettingLayout->addWidget(consoleTimeCheck, 1, 0, 1, 1);

    mainSettingWidget->setLayout(mainSettingLayout);

    /////////////// Tables

    QWidget*     tablesWidget = new QWidget(this);
    QGridLayout* tablesLayout = new QGridLayout(tablesWidget);

    QLabel* tableLabel = new QLabel(tablesWidget);
    tableLabel->setText("Table: ");

    QComboBox* tableCombo = new QComboBox(tablesWidget);
    tableCombo->addItem("Sessions");

    tablesLayout->addWidget(tableLabel, 0, 0, 1, 1);
    tablesLayout->addWidget(tableCombo, 0, 1, 1, 1);

    tablesWidget->setLayout(tablesLayout);

    //////////////

    listSettings = new QListWidget(this);
    listSettings->setFixedWidth(150);
    listSettings->setAlternatingRowColors(true);
    listSettings->addItem("Main settings");
    listSettings->addItem("Tables");

    labelHeader = new QLabel(this);
    labelHeader->setStyleSheet("font-size: 16px; font-weight: bold; margin-bottom: 10px;");
    labelHeader->setText("Main settings");

    lineFrame = new QFrame(this);
    lineFrame->setFrameShape(QFrame::HLine);
    lineFrame->setFrameShadow(QFrame::Plain);

    headerLayout = new QVBoxLayout();
    headerLayout->addWidget(labelHeader);
    headerLayout->addWidget(lineFrame);
    headerLayout->setSpacing(0);

    hSpacer      = new QSpacerItem(40, 20, QSizePolicy::Expanding, QSizePolicy::Minimum);
    buttonApply     = new QPushButton("Apply ", this);
    buttonClose = new QPushButton("Cancel", this);

    stackSettings = new QStackedWidget(this);
    stackSettings->addWidget(mainSettingWidget);
    stackSettings->addWidget(tablesWidget);

    layoutMain = new QGridLayout(this);
    layoutMain->setContentsMargins(4, 4, 4, 4);
    layoutMain->addWidget(listSettings, 0, 0, 2, 1);
    layoutMain->addLayout(headerLayout, 0, 1, 1, 3);
    layoutMain->addWidget(stackSettings, 1, 1, 1, 3);
    layoutMain->addItem(hSpacer, 2, 1, 1, 1);
    layoutMain->addWidget(buttonApply, 2, 2, 1, 1);
    layoutMain->addWidget(buttonClose, 2, 3, 1, 1);

    this->setLayout(layoutMain);
}

void DialogSettings::onStackChange(int index)
{
    QString text = listSettings->item(index)->text();
    labelHeader->setText(text);
    stackSettings->setCurrentIndex(index);
}

void DialogSettings::onApply()
{

}