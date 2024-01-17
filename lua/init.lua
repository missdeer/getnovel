-- NovelChapterInfo 结构体的模拟
function NovelChapterInfo(index, title, url)
    return { Index = index, Title = title, URL = url }
end

-- 定义 GetNovelSiteExternalHandler 类型
GetNovelSiteExternalHandler = {}
GetNovelSiteExternalHandler.__index = GetNovelSiteExternalHandler

-- 构造函数
function GetNovelSiteExternalHandler:new(title, urls, canHandleFunc, preprocessChapterListURLFunc, extractChapterListFunc, extractChapterContentFunc)
    local instance = {}
    setmetatable(instance, GetNovelSiteExternalHandler)
    
    instance.Title = title
    instance.Urls = urls
    instance.CanHandle = canHandleFunc
    instance.PreprocessChapterListURL = preprocessChapterListURLFunc
    instance.ExtractChapterList = extractChapterListFunc
    instance.ExtractChapterContent = extractChapterContentFunc

    return instance
end

-- 创建 handlers 全局数组
local handlers = {}

-- 全局变量，用于存储当前处理器
local currentHandler = nil

-- RegisterHandler 注册处理器
function RegisterHandler(handler)
    table.insert(handlers, handler)
end

-- FindHandler 查找可处理指定 URL 的处理器
function FindHandler(url)
    for _, handler in ipairs(handlers) do
        if handler.CanHandle(url) then
            currentHandler = handler
            return true
        end
    end
    return false
end

-- 定义 PreprocessChapterListURL 预处理章节列表 URL，其实就是根据 URL 的特征，将其转换为章节列表的 URL
function PreprocessChapterListURL(url)
    if currentHandler ~= nil then
        return currentHandler.PreprocessChapterListURL(url)
    end
end

-- 定义 ExtractChapterList 从url中提取章节列表
function ExtractChapterList(url, rawPageContent)
    if currentHandler ~= nil then
        return currentHandler.ExtractChapterList(url, rawPageContent)
    end
end

-- 定义 ExtractChapterContent 清洗章节内容，比如去除一些干扰用的字符串等，仅保留正文内容
function ExtractChapterContent(rawPageContent)
    if currentHandler ~= nil then
        return currentHandler.ExtractChapterContent(rawPageContent)
    end
end
