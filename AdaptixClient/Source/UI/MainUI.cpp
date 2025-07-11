#include <Agent/Agent.h>
#include <UI/MainUI.h>
#include <UI/Widgets/AdaptixWidget.h>
#include <UI/Widgets/SessionsTableWidget.h>
#include <UI/Widgets/TasksWidget.h>
#include <UI/Graph/SessionsGraph.h>
#include <UI/Dialogs/DialogExtender.h>
#include <UI/Dialogs/DialogSettings.h>
#include <Client/Extender.h>
#include <Client/Settings.h>
#include <Client/AuthProfile.h>
#include <MainAdaptix.h>

MainUI::MainUI()
{
    this->setWindowTitle( FRAMEWORK_VERSION );

    auto newProjectAction = new QAction("New Project", this);
    connect(newProjectAction, &QAction::triggered, this, &MainUI::onNewProject);
    auto closeProjectAction = new QAction("Close Project", this);
    connect(closeProjectAction, &QAction::triggered, this, &MainUI::onCloseProject);

    auto menuProject = new QMenu("Projects", this);
    menuProject->addAction(newProjectAction);
    menuProject->addAction(closeProjectAction);

    auto menuExtender = new QMenu("Extender", this);
    auto extenderAction = new QAction("Open extender", this);
    connect(extenderAction, &QAction::triggered, this, &MainUI::onExtender);
    menuExtender->addAction(extenderAction);

    auto menuSettings = new QMenu("Settings", this);
    auto settingsAction = new QAction("Open settings", this);
    connect(settingsAction, &QAction::triggered, this, &MainUI::onSettings);
    menuSettings->addAction(settingsAction);

    auto mainMenuBar = new QMenuBar(this);
    mainMenuBar->addMenu(menuProject);
    mainMenuBar->addMenu(menuExtender);
    mainMenuBar->addMenu(menuSettings);

    this->setMenuBar(mainMenuBar);

    mainuiTabWidget = new QTabWidget();
    mainuiTabWidget->setTabPosition(QTabWidget::South);
    mainuiTabWidget->tabBar()->setMovable(true);
    mainuiTabWidget->setMovable(true);

    this->setCentralWidget(mainuiTabWidget);
}

MainUI::~MainUI()
{
    for (auto project : AdaptixProjects.keys()) {
        delete AdaptixProjects[project];
        AdaptixProjects.remove(project);
    }
}

void MainUI::closeEvent(QCloseEvent* event)
{
    QCoreApplication::quit();
    event->accept();
}

void MainUI::AddNewProject(AuthProfile* profile, QThread* channelThread, WebSocketWorker* channelWsWorker)
{
    auto adaptixWidget = new AdaptixWidget(profile, channelThread, channelWsWorker);
    if (!adaptixWidget)
        return;

    for (auto extFile : GlobalClient->extender->extenderFiles){
        if(extFile.Valid && extFile.Enabled)
            adaptixWidget->AddExtension(extFile);
    }

    QString tabName = "   " + profile->GetProject() + "   " ;
    int id = mainuiTabWidget->addTab( adaptixWidget, tabName);
    mainuiTabWidget->setCurrentIndex( id );

    AdaptixProjects[profile->GetProject()] = adaptixWidget;
}

void MainUI::AddNewExtension(const ExtensionFile &extFile)
{
    for (auto adaptixWidget : AdaptixProjects) {
        if (adaptixWidget)
            adaptixWidget->AddExtension(extFile);
    }
}

void MainUI::RemoveExtension(const ExtensionFile &extFile)
{
    for (auto adaptixWidget : AdaptixProjects) {
        if (adaptixWidget)
            adaptixWidget->RemoveExtension(extFile);
    }
}

void MainUI::UpdateSessionsTableColumns() {
    for (auto adaptixWidget : AdaptixProjects) {
        if (adaptixWidget)
            adaptixWidget->SessionsTablePage->UpdateColumnsVisible();
    }
}

void MainUI::UpdateGraphIcons() {
    for (auto adaptixWidget : AdaptixProjects) {
        if (adaptixWidget) {
            for (auto agent : adaptixWidget->AgentsMap.values() ) {
                agent->UpdateImage();
            }
            adaptixWidget->SessionsGraphPage->UpdateIcons();
        }
    }
}

void MainUI::UpdateTasksTableColumns() {
    for (auto adaptixWidget : AdaptixProjects) {
        if (adaptixWidget)
            adaptixWidget->TasksTab->UpdateColumnsVisible();
    }
}

/// Actions

void MainUI::onNewProject()
{
    GlobalClient->NewProject();
}

void MainUI::onCloseProject()
{
    int currentIndex = mainuiTabWidget->currentIndex();
    auto adaptixWidget = qobject_cast<AdaptixWidget*>( mainuiTabWidget->currentWidget() );
    if (!adaptixWidget)
        return;

    if (adaptixWidget) {
        AdaptixProjects.remove(adaptixWidget->GetProfile()->GetProject());
        adaptixWidget->Close();
        delete adaptixWidget;
    }
    mainuiTabWidget->removeTab(currentIndex);
}

void MainUI::onExtender()
{
    GlobalClient->extender->dialogExtender->show();
}

void MainUI::onSettings()
{
    GlobalClient->settings->dialogSettings->show();
}
