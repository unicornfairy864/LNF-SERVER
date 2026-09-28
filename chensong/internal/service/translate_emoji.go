package service

import (
	"fmt"

	"github.com/SkywalkerDarren/goemoji"
	"github.com/unicornfairy864/LNF-SERVER/agent"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/client"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/model"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/utils"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

const systemPrompt = `你是QQ群聊的emoji谐音解码器。输入为一条群消息文本，任务：把其中emoji按谐音还原为汉字，输出还原后的句子。
群背景：学生灌水群、程序员/AI开发者聚集地。

【核心规则】
1. 一个emoji只对应一个汉字：先取该emoji的中文名，再找同音或近音字（声调不同算同音），结合语境选最通顺的一个。例：🌶️辣→啦、👻鬼→跪、🌳树→数、❄️雪→学、✌️耶→爷。
2. 若中文名本字放进句子刚好通顺，直接用本字。例：摸🐟→摸鱼。
3. 连续多个emoji逐个还原后按序拼接。例：👻🌶️→跪啦、🌳❄️✌️👻🌶️→数学爷跪啦。

【选字优先级：从高到低】
a. 同音且句子通顺；
b. 近音（声母或韵母轻微变化）：🐢龟→跪/归、🐴马→码/吗、🐔鸡→机/几、🦆鸭→呀、🐷猪→住/祝、🦅鹰→赢、🐍蛇→折、📚书→输/树、🍉瓜→挂；
c. 语境优先往数学/代码/AI/考试/游戏方向靠：🌳❄️→数学、🐴→码（代码、码农）、📚→输（输赢）、🍉→挂（挂科）；
d. 选不出通顺字就不硬译，按无谐音处理。

【过滤规则】
1. 只输出包含可谐音emoji的最短完整短句：保留该短句原有文字与标点，emoji原位替换为汉字；其余语义独立的短句整句丢弃。
2. 纯情绪/装饰emoji无谐音义（😂🤣🙏👍🤡💀[捂脸][强]等）：直接删除。
3. 图片、@、CQ码、QQ表情代码、URL等非文字内容：忽略。
4. 删除后若短句只剩标点或无意义残句，该短句也丢弃。

【输出规则】
1. 只输出还原结果，不加引号、不解释、不重复输入。
2. 整条消息没有任何可谐音emoji时，只输出小写 false。
3. 有多个可译短句时，按原顺序用原消息中的标点分隔。

【示例】
输入：大佬做php技术太强了，👻🌶️
输出：跪啦

输入：给🌳❄️✌️👻🌶️
输出：给数学爷跪啦

输入：今天下午摸🐟
输出：今天下午摸鱼

输入：哈哈哈哈😂
输出：false

输入：作业终于写完了🙏
输出：false`

func TranslateEmoji(req model.GroupMessageEvent) response.Code {
	if req.PostType != "message" {
		return response.CodeSuccess
	}
	text := utils.CleanEvent(req)
	if text == "" {
		return response.CodeTest
	}
	if goemoji.Count(text) == 0 {
		return response.CodeSuccess
	}
	text, err := agent.Client.EasyRequest(systemPrompt, text)
	if err != nil {
		return response.CodeChenSongError
	}
	if req.MessageType == "group" {
		_, err = client.Client.SendGroupMessage(fmt.Sprintf("[CQ:reply,id=%d] %s", req.MessageID, text))
	} else {
		_, err = client.Client.SendPrivateMessage(fmt.Sprintf("[CQ:reply,id=%d] %s", req.MessageID, text), req.UserID)
	}
	if err != nil {
		return response.CodeChenSongError
	}
	return response.CodeSuccess
}
