#include <Utils/FontManager.h>
#include <QDebug>
#include <QFontInfo>

FontManager& FontManager::instance()
{
    static FontManager instance;
    return instance;
}

void FontManager::initialize()
{
    if (m_initialized) {
        return;
    }

    loadApplicationFonts();
    m_initialized = true;
}

void FontManager::loadApplicationFonts()
{
    struct FontResource {
        QString resourcePath;
        QString alias;
    };

    QList<FontResource> fonts = {
        {":/fonts/Hack", "Hack"},
        {":/fonts/Hack_B", "Hack"},
        {":/fonts/Hack_BI", "Hack"},
        {":/fonts/Hack_I", "Hack"},
        {":/fonts/JetBrainsMono", "JetBrains Mono"},
        {":/fonts/JetBrainsMono_B", "JetBrains Mono"},
        {":/fonts/JetBrainsMono_BI", "JetBrains Mono"},
        {":/fonts/JetBrainsMono_I", "JetBrains Mono"}
    };

    for (const auto& fontRes : fonts) {
        int fontId = QFontDatabase::addApplicationFont(fontRes.resourcePath);
        if (fontId != -1) {
            QStringList fontFamilies = QFontDatabase::applicationFontFamilies(fontId);
            if (!fontFamilies.isEmpty()) {
                QString actualFamilyName = fontFamilies.first();
                m_loadedFonts[fontRes.alias] = actualFamilyName;
            }
        }
    }
}

QFont FontManager::getFont(const QString& fontName, int pointSize)
{
    if (!m_initialized)
        initialize();

    QFont font;

    if (m_loadedFonts.contains(fontName)) {
        font = QFont(m_loadedFonts[fontName]);
    } else {
        font = getDefaultMonospaceFont();
    }

    if (pointSize > 0)
        font.setPointSize(pointSize);

    return font;
}

bool FontManager::isFontAvailable(const QString& fontName)
{
    if (!m_initialized)
        initialize();

    return m_loadedFonts.contains(fontName);
}

QFont FontManager::getDefaultMonospaceFont(int pointSize)
{
    if (!m_initialized)
        initialize();

    if (m_loadedFonts.contains("JetBrains Mono")) {
        QFont font(m_loadedFonts["JetBrains Mono"]);
        if (pointSize > 0)
            font.setPointSize(pointSize);
        return font;
    }

    if (m_loadedFonts.contains("Hack")) {
        QFont font(m_loadedFonts["Hack"]);
        if (pointSize > 0)
            font.setPointSize(pointSize);
        return font;
    }

    QFont font;
    font.setFamily("monospace");
    font.setStyleHint(QFont::Monospace);
    if (pointSize > 0)
        font.setPointSize(pointSize);

    return font;
}

QString FontManager::findBestMonospaceFont()
{
    if (!m_initialized)
        initialize();

    if (m_loadedFonts.contains("JetBrains Mono"))
        return "JetBrains Mono";

    if (m_loadedFonts.contains("Hack"))
        return "Hack";

    return "monospace";
}