import type { ImageStudioMode } from '@/api/imageStudio'

export type ImageStudioPromptGalleryRatio =
  | '1:1'
  | '3:2'
  | '2:3'
  | '4:3'
  | '3:4'
  | '5:4'
  | '4:5'
  | '16:9'
  | '9:16'
  | '21:9'

export type ImageStudioPromptGalleryResolution = '1K' | '2K' | '4K'
export type ImageStudioPromptGalleryQuality = 'high' | 'medium' | 'low' | 'auto'

export interface ImageStudioPromptGalleryCategory {
  id: string
  name: string
  description: string
  accent: string
}

export interface ImageStudioPromptGalleryItem {
  id: string
  categoryId: string
  title: string
  mode: ImageStudioMode
  prompt: string
  ratio: ImageStudioPromptGalleryRatio
  resolution: ImageStudioPromptGalleryResolution
  quality: ImageStudioPromptGalleryQuality
  count: 1 | 2 | 3 | 4
  image: string
  tags: string[]
}

type GallerySeed = Omit<ImageStudioPromptGalleryItem, 'image'> & {
  image?: string
  palette: [string, string, string]
  visual: string
}

export const imageStudioPromptGalleryCategories: ImageStudioPromptGalleryCategory[] = [
  { id: 'portrait', name: '人像写真', description: '自然光、棚拍、情绪写真', accent: '#2563eb' },
  { id: 'graduation', name: '证件/毕业照', description: '正式合影、证件形象、年代感', accent: '#0f766e' },
  { id: 'architecture', name: '建筑城市', description: '外立面、城市街景、空间叙事', accent: '#475569' },
  { id: 'illustration', name: '插画绘本', description: '儿童绘本、 editorial、治愈插画', accent: '#16a34a' },
  { id: 'anime', name: '动漫角色', description: '角色设定、场景氛围、头像', accent: '#7c3aed' },
  { id: 'shooting', name: '拍摄', description: '胶片、宝丽来、摄影质感', accent: '#0891b2' },
]

const seeds: GallerySeed[] = [
  {
    id: 'portrait-window-light',
    categoryId: 'portrait',
    title: '窗边自然光肖像',
    mode: 'text_to_image',
    prompt: '年轻女性半身人像，站在浅色窗帘旁，柔和自然光从侧面照入，神情安静自信，皮肤质感真实，背景简洁，高级杂志写真风格，85mm 镜头，浅景深，真实摄影，高细节。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['自然光', '写真', '浅景深'],
    palette: ['#dbeafe', '#f8fafc', '#1d4ed8'],
    visual: 'portrait'
  },
  {
    id: 'portrait-business-headshot',
    categoryId: 'portrait',
    title: '商务头像形象照',
    mode: 'image_to_image',
    prompt: '基于参考人物生成专业商务头像，保持五官和身份特征一致，深灰西装与白衬衫，干净浅灰背景，柔和棚拍布光，表情自然可信，适合个人主页和商务资料。',
    ratio: '1:1',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '头像', '商务'],
    palette: ['#e5e7eb', '#ffffff', '#111827'],
    visual: 'portrait'
  },
  {
    id: 'portrait-cinematic-night',
    categoryId: 'portrait',
    title: '电影感夜景人像',
    mode: 'text_to_image',
    prompt: '夜晚城市街头人像，霓虹灯在湿润路面形成倒影，人物穿深色风衣，回头看向镜头，蓝紫和暖橙对比光，电影剧照质感，真实摄影，细节丰富。',
    ratio: '2:3',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['电影感', '夜景', '霓虹'],
    palette: ['#312e81', '#f97316', '#020617'],
    visual: 'portrait'
  },
  {
    id: 'portrait-summer-cafe',
    categoryId: 'portrait',
    title: '夏日咖啡馆写真',
    mode: 'text_to_image',
    prompt: '夏日下午咖啡馆外的生活方式人像，人物坐在木质桌边，白衬衫与浅色牛仔裤，桌上有冰咖啡和书本，阳光穿过树影，轻松自然，日系胶片色调。',
    ratio: '4:5',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['生活方式', '日系', '胶片'],
    palette: ['#fef3c7', '#86efac', '#92400e'],
    visual: 'portrait'
  },
  {
    id: 'portrait-chinese-wedding-bouquet',
    categoryId: 'portrait',
    title: '中式婚纱手捧花',
    mode: 'image_to_image',
    prompt: '参考上传的图片，在严格保持图中人物的身材、相貌特征、面部五官、肤色、发型、表情气质和身材比例一致的前提下，重新拍摄一张中式婚纱人像照片。人物身穿精致中式婚纱或中式婚礼礼服，服装质感高级、刺绣细节清晰自然，手捧鲜花并看向镜头，姿态端庄自然，神情温柔自信。画面为真实婚纱摄影风格，竖幅构图，半身或三分之二身人像，背景简洁干净，柔和棚拍光或温暖自然光，皮肤质感真实，细节清晰，高级、浪漫、典雅。不要改变人物身份、脸型、五官、肤色、发型和身体比例，不要卡通化、不要过度磨皮、不要塑料感、不要脸部变形、不要服装廉价感、不要背景杂乱。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '中式婚纱', '手捧花'],
    palette: ['#fff1f2', '#dc2626', '#7f1d1d'],
    visual: 'portrait'
  },
  {
    id: 'portrait-q-version-chinese-wedding',
    categoryId: 'portrait',
    title: 'Q版中式古装婚礼',
    mode: 'image_to_image',
    prompt: '参考上传的照片，将照片里的两个人转换成 Q 版 3D 人物，保持两个人各自的核心面部特征、发型、肤色、表情气质和身份辨识度一致。整体场景为中式古装婚礼，大红喜庆色调，背景使用“囍”字剪纸风格图案，氛围热烈、吉祥、华丽。服饰写实精致：男士身着中式长袍马褂，主体为红色，上面有金色绣龙纹图案，彰显尊贵大气，胸前系着大红花，寓意喜庆吉祥；女士身穿秀禾服，同样以红色为基调，饰有精美金色花纹与凤凰刺绣，展现典雅华丽之感，头上搭配花朵发饰，增添柔美温婉气质。二者皆为中式婚礼经典着装，蕴含对新人婚姻美满的祝福。头饰要求：男士佩戴中式状元帽，主体红色，饰有金色纹样，帽顶有精致金饰，传统儒雅庄重；女士佩戴凤冠造型，以红色花朵为中心，搭配金色立体装饰与垂坠流苏，华丽富贵，古典韵味十足。画面为高质量 3D 卡通渲染，Q 版比例可爱但五官辨识度清晰，材质细腻，红金配色高级，构图端正完整，人物自然并肩站立或亲密靠近，看向镜头，喜庆浪漫。避免脸部不一致、五官丢失、身体畸形、廉价塑料感、服装细节混乱、文字乱码、背景杂乱、过度夸张变形。',
    ratio: '1:1',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', 'Q版3D', '中式婚礼'],
    palette: ['#fee2e2', '#dc2626', '#facc15'],
    visual: 'portrait'
  },
  {
    id: 'portrait-western-wedding-forehead-touch',
    categoryId: 'portrait',
    title: '西式婚纱额头相贴',
    mode: 'image_to_image',
    prompt: '严格参考上传照片中的情侣角色，保持两个人的面部五官、肤色、发型、表情气质、身材比例和身份辨识度一致。在此前提下生成一张西式婚纱影楼拍摄风格照片：男士身穿正式西装，女士身穿白色婚纱，整体造型精致自然。情侣额头轻轻相贴，闭眼微笑，动作亲密温柔，氛围浪漫温馨。画面使用柔和影楼灯光，浅景深，背景简洁干净，半身或全身构图均可，高分辨率，细节清晰，真实皮肤质感，婚纱面料和西装质感自然高级。避免脸部不一致、五官变形、身材比例改变、服装变形、过度磨皮、塑料感、背景杂乱、廉价影楼感、卡通化、AI 痕迹。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '西式婚纱', '情侣'],
    palette: ['#f8fafc', '#f3e8ff', '#475569'],
    visual: 'portrait'
  },
  {
    id: 'portrait-chinese-wedding-shoulder-embrace',
    categoryId: 'portrait',
    title: '中式婚纱靠肩拥抱',
    mode: 'image_to_image',
    prompt: '严格参考上传照片中的情侣形象，保持两个人的面部五官、肤色、发色、发型、表情气质、身材比例和身份辨识度一致。在此前提下生成一张中式婚纱影楼风格照片：男士身穿红色中式礼服，女士身穿红色旗袍或凤冠霞帔，服饰质感高级，红色主调喜庆典雅，刺绣和金色纹样细节清晰自然。场景为古典中式布景，可包含中式屏风、红色帘幕、传统纹样、柔和花艺或喜庆装饰，整体干净高级。男士轻轻拥着女士肩膀，女士微微靠向男士，动作自然亲密，浪漫温馨。画面使用柔和影楼灯光，浅景深，半身或全身构图均可，高分辨率，细节清晰，真实皮肤质感，中式婚礼氛围浓郁。避免脸部不一致、五官变形、身材比例改变、服装变形、廉价影楼感、背景杂乱、过度磨皮、塑料感、卡通化、AI 痕迹。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '中式婚纱', '靠肩拥抱'],
    palette: ['#fee2e2', '#b91c1c', '#facc15'],
    visual: 'portrait'
  },
  {
    id: 'portrait-couple-polaroid-booth',
    categoryId: 'portrait',
    title: '情侣相片亭拼贴',
    mode: 'image_to_image',
    prompt: '根据上传的两张照片，严格保持两个人的面部五官、肤色、发型、表情、身材比例 100% 相同且完全不做任何风格化修改，生成一张复古风格拍立得/相片亭多帧竖向拼贴大片。画面呈现同一对情侣在相片亭里的多个自然、亲密、俏皮瞬间。拼贴包含以下场景，每帧一张，共 6-8 张自然排列：女人灿烂微笑，男人站在她身后双手轻轻遮住她眼睛，俏皮玩闹；两人面对面站得很近，女人手轻放在男人胸口，深情对视；两人脸贴脸，柔和微笑，眼神交汇；女人站在男人身后，对镜头比出和平手势，露出甜美笑容；俏皮跳舞姿势，男人单手拉起女人一只手，像在旋转她；两人自然大笑，对镜头放松真实瞬间。环境为经典相片亭中性背景，柔和垂直布帘，温暖室内光线，微弱阴影，亲密温馨氛围。摄影风格为复古拍立得/胶片效果，轻微柔焦、细腻颗粒、轻微模糊、自然瑕疵、暖色调、真实皮肤质感。灯光为柔和正面闪光灯与环境光结合，营造真实生活快照感。整体氛围浪漫、俏皮、即兴、亲密、青春、怀旧。构图为竖向多帧拼贴，帧间均匀间隔，真实相片亭布局，高真实度，无 AI 痕迹，自然比例，无畸变。避免出现：面部变形、脸部不一致、身材改变、过度模糊、色彩失真、卡通风格、塑料感、廉价感、过度修图、背景杂乱、服装改变、阴影过重、现代数码感。',
    ratio: '9:16',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '情侣', '相片亭'],
    palette: ['#fef2f2', '#f59e0b', '#7f1d1d'],
    visual: 'portrait'
  },
  {
    id: 'portrait-shopping-bag-upshot-ad',
    categoryId: 'portrait',
    title: '购物袋仰拍广告',
    mode: 'image_to_image',
    prompt: '从购物袋内部向上拍摄一张充满活力、可爱、高端的广告照片，通过袋子顶部的大开口朝向天空。相机放置在购物袋深处，使用 f/1.4 广角镜头，营造俏皮的特写视角，带有轻微但迷人的广角失真。这种失真必须清晰保留，赋予图像有趣、沉浸式的顶级广告感。购物袋边缘自然框定图像，展示高端纸张质感、柔和褶皱、手柄以及干净的商业包装美学。购物袋采用高饱和度颜色，可随机选择热粉色、饱和黄色、血橙色、钴蓝色、鲜艳绿色、电紫色、深青色或其他大胆商业色彩，整体明亮、欢快、多彩且具有视觉冲击力。主要主体：如果上传了 FACE_REF，严格以 100% 保真度保留上传人物的面部身份，保持精确的面部结构、眼睛、眼睑、鼻子、嘴唇、下巴线、颧骨、皮肤纹理、年龄印象、发型倾向以及整体身份，不得进行美化漂移、面部重塑、年龄改变、种族改变或虚构特征。如果仅提供 PERSON_NAME，则基于 [周星驰] 的公众形象，创作一幅高度可识别的肖像，尽可能忠实保留其标志性面部特征、发型、表情倾向、姿势以及个人气质。人物从上方深深探入购物袋，将头部完全浸入开口，仿佛好奇地发现袋子里的东西。人物的脸非常靠近相机镜头，直视镜头，拥有睁大的、富有灵魂的、惹人喜爱的眼睛和温暖可爱的表情。这种姿势应俏皮、亲密、心动，仿佛人物好奇地窥探袋子并意外与观众眼神接触。光线明亮、自然且振奋人心，来自上方天空的柔和阳光。图像最大限度展现可爱、魅力和“哇”因素，同时仍像一张精致的奢侈广告照片，最终结果应立即让观众微笑。避免脸部不一致、五官变形、身份漂移、过度美化、年龄改变、广角失真过度导致畸形、购物袋边缘缺失、包装廉价感、背景杂乱、塑料感、AI 痕迹。',
    ratio: '4:5',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '广告', '仰拍'],
    palette: ['#fdf2f8', '#ec4899', '#0ea5e9'],
    visual: 'portrait'
  },
  {
    id: 'portrait-nostalgic-classroom-cinema',
    categoryId: 'portrait',
    title: '怀旧肖像',
    mode: 'image_to_image',
    prompt: '根据上传图片中的男子生成电影式肖像，严格保留面部细节 100%，保持五官、脸型、肤色、发型、表情气质、身材比例和身份辨识度一致。场景为复古教室，镜头从远处拍摄，椅背的一小部分在前景中清晰可见，仿佛摄影师正在谨慎地捕捉这幅肖像，或捕捉一个静谧时刻。柔和的金色光线从右侧墙壁上的小窗户或开口射入，形成对角线，直射在男子脸上，营造戏剧性和情感对比。男子独自坐在椅子上，双脚搁在桌子上，表情轻松冷峻，仿佛陷入沉思。男子身穿宽松的黑色军装风格毛衣，搭配奶油色工装裤、匡威运动鞋和红色耳机。椅子上挂着一个包，与周围温暖光线和谐融合。身后是一面灰白色墙壁，上面贴满标有“beingbb”、作业安排提醒等内容的便利贴，顶部挂着一张照片，并有大学教室常见的装饰或家具。强化怀旧与沉思氛围，仿佛时光凝固。柔和金色光线与房间暗影交相辉映，营造平静、温暖而略带忧郁的氛围，令人联想到日本独立电影中黄昏或清晨的场景。所有视觉元素都不要呈现散景，从前景到背景保持均匀锐度。视觉纹理包含明显噪点和颗粒感，类似佳能 AE-1 等 35 毫米模拟相机效果，或富士 X100V 搭配“经典正片”胶片模拟的复古数字模拟效果。可选相机设置：ISO 1600、光圈 f/5.6、快门速度 1/60 秒，暖白平衡，保留房间内自然金色光线。颗粒效果可来自 ISO 400 胶片或数字颗粒功能，增强电影感与怀旧感。避免脸部不一致、五官变形、身材比例改变、过度磨皮、现代教室感、背景杂乱、散景虚化、过度锐化、塑料感、AI 痕迹。',
    ratio: '4:3',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '怀旧', '电影感'],
    palette: ['#fef3c7', '#92400e', '#111827'],
    visual: 'portrait'
  },
  {
    id: 'portrait-child-haimati-studio',
    categoryId: 'portrait',
    title: '宝宝海马体写真',
    mode: 'image_to_image',
    prompt: '将上传的儿童照片转化为高端影楼写真，参考海马体/天真蓝风格。保持孩子原有五官与身份特征，可微调表情至自然微笑或轻松愉快状态。替换背景为摄影棚纯色浅调，如米白、奶油色或浅灰，搭配柔和棚拍灯光，画面简洁无杂物。升级服装为质感影款，例如纯色毛衣、简约连衣裙或小针织，色彩柔和协调，避免花纹与艳色。调整姿态至自然童趣动作，例如微侧身、轻靠或抬手，放松不生硬。整体色调明亮柔和，突出高级感与童真，保留皮肤与衣物细节，拒绝过度美颜。采用 3:4 竖版特写或半身构图，人物居中或微偏，适当留白。输出高清高质感人像摄影。避免五官变形、身份特征丢失、成人化妆造型、过度磨皮、皮肤塑料感、背景杂乱、服装艳俗、表情僵硬、AI 痕迹。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '宝宝写真', '影楼'],
    palette: ['#fff7ed', '#f5e7d0', '#94a3b8'],
    visual: 'portrait'
  },
  {
    id: 'portrait-child-zhou-catching',
    categoryId: 'portrait',
    title: '抓周',
    mode: 'image_to_image',
    prompt: '参考上传的宝宝照片，保持宝宝面部特征、发色、肤色、体型、年龄感和身份辨识度完全一致，生成满周岁抓周照片。整体为中国传统风格，宝宝坐在红色或喜庆布料铺成的抓周席上，周围摆放传统抓周道具，道具可随机包含毛笔、算盘、印章、书本、钱币等。宝宝正伸手抓其中一个道具，表情好奇、专注或微笑，动作自然可爱。光线柔和自然，温暖色调，高分辨率，浅景深，儿童摄影风格，可采用俯视或正面拍摄，背景柔和布置，突出传统节日氛围。避免五官变形、身份特征丢失、体型改变、成人化、道具杂乱、背景廉价、过度磨皮、皮肤塑料感、表情僵硬、AI 痕迹。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '宝宝', '抓周'],
    palette: ['#fee2e2', '#dc2626', '#f59e0b'],
    visual: 'portrait'
  },
  {
    id: 'shooting-polaroid-style',
    categoryId: 'shooting',
    title: '宝丽来风格',
    mode: 'image_to_image',
    prompt: '结合上传的图片，保持图片中人物或主体的形象特征、五官比例、肤色、发型、服装轮廓和整体辨识度一致，将其转化成拍立得风格照片。图像呈现怀旧的胶卷照片风格，具有明显的胶卷颗粒、轻微的运动模糊和柔和的焦点。画面中带有微妙的色彩闪烁，以及紫色、粉色、青绿色的光线泄漏，赋予照片梦幻般的复古宝丽来感觉。照片边框像经典宝丽来相纸，四周有白色边框，底部有宽宽的白色边框。整体为真实生活快照质感，暖色调，轻微柔焦，自然瑕疵，复古、浪漫、随性但主体清晰。避免人物面部变形、身份特征丢失、过度模糊、色彩脏乱、边框缺失、现代数码感、塑料感、卡通化、背景杂乱。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '宝丽来', '胶片'],
    palette: ['#ecfeff', '#0891b2', '#a855f7'],
    visual: 'retro'
  },
  {
    id: 'shooting-cinematic-portrait',
    categoryId: 'shooting',
    title: '电影级肖像',
    mode: 'image_to_image',
    prompt: '根据参考图角色生成一位五官、外貌、身材、肤色、发型特征和整体身份辨识度完全一致的人物形象照。画面为梦幻氛围，柔和阳光在人物脸上投下精致阴影，浅景深，超现实但真实可信的皮肤纹理，温暖电影色调，极简美学，时尚编辑摄影质感，85mm 镜头，背景虚化，高细节，柔和对比，自然光线，情感人像构图，独立电影静止画面氛围，逼真写实，4K。人物为短发凌乱的波波头，柔和自然妆容，身穿超大号蓝色卫衣，站在晴朗蓝天下金色时段的户外近景电影人像。风吹乱头发，表情忧郁，镜头捕捉自然真实的情绪瞬间。避免脸部不一致、五官变形、身材比例改变、过度磨皮、塑料感、卡通化、背景杂乱、廉价写真感、AI 痕迹。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '电影感', '肖像'],
    palette: ['#dbeafe', '#facc15', '#1e3a8a'],
    visual: 'portrait'
  },
  {
    id: 'shooting-first-person-iphone',
    categoryId: 'shooting',
    title: '第一视角摄影',
    mode: 'image_to_image',
    prompt: '结合上传的图片人物形象和特征，保持人物的面部五官、肤色、发型、表情气质、身材比例和身份辨识度一致，转成一张极其平庸的 iPhone 照片。画面中人物面对镜头微笑，拍摄者第一视角、平视角度牵着图中人物的手。由于阳光不均匀，画面略微过曝，光线有些生硬。角度尴尬，构图拙劣，整体效果平庸至极，像普通人随手拍的生活照。自然而然地竖屏拍摄，真实手机照片质感，不要专业大片感，不要精修，不要过度美化。避免人物脸部不一致、五官变形、身材比例改变、过度清晰棚拍感、过度电影感、卡通化、塑料感、背景过度设计、AI 痕迹。',
    ratio: '9:16',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '第一视角', '手机随拍'],
    palette: ['#fef3c7', '#38bdf8', '#64748b'],
    visual: 'portrait'
  },
  {
    id: 'shooting-flash-street-girl',
    categoryId: 'shooting',
    title: '少女感街拍',
    mode: 'image_to_image',
    prompt: '参考上传照片中的年轻成年女性，保持人物五官、脸型、笑容、肤色、发型、发色、身材比例和整体身份辨识度完全一致，但双眼必须完全睁开，明亮有神，水汪汪地直视镜头，绝不眯眼。画面为超写实街拍风格，真实自然光与直射闪光灯混合照明，强烈闪光灯直打在脸上，高光爆亮，皮肤反光真实。一位阳光开朗的年轻女性站在人行道上，露出超级灿烂自信的笑容，双眼大睁明亮有神，左手比出超萌 V 字手势，右手举着一部智能手机，手机屏幕清晰显示她本人自拍，屏幕里的她也笑容满面、双眼睁开，头发带有粉色花朵贴纸滤镜效果。妆容为浓密卷翘睫毛、水润少女妆，脸颊和鼻尖有淡淡粉色腮红，渐变粉色水光唇，皮肤在闪光灯下水润透亮。背景为阳光下的城市街头、商店橱窗和水泥墙，微微虚化。整体氛围青春元气、ins 风、韩国女团同款活力、闪光灯街拍感、照片级真实。头部佩戴棕色渔夫帽，帽子上别着几枚蓝色徽章，发型和头发颜色保持与上传照片一致。服装为白色拉链露脐短上衣、浅蓝色高腰宽松牛仔裤、棕色皮带。配饰包含左手腕超粗银色链条手链、红白条纹手提包挂蓝色鲨鱼钥匙扣、腰带环上挂红色史努比零钱包。手机和头部周围漂浮色彩缤纷的星星、爱心、闪光贴纸特效，可爱梦幻。8K 质感，超高清，顶级写实，皮肤毛孔可见，闪光灯高光真实，细节丰富，锐利。避免脸部不一致、五官变形、眯眼、闭眼、身材比例改变、过度磨皮、塑料感、廉价滤镜感、背景杂乱、服装配饰缺失、手机屏幕错误、AI 痕迹。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '闪光灯', '街拍'],
    palette: ['#fdf2f8', '#ec4899', '#38bdf8'],
    visual: 'portrait'
  },
  {
    id: 'shooting-kawaii-food-collage',
    categoryId: 'shooting',
    title: '可爱美食拼贴',
    mode: 'text_to_image',
    prompt: '可爱日式卡哇伊风格美食摄影拼贴，温馨温暖美学，从上往下俯拍咖啡馆桌上的舒适美食和饮品，包含汉堡、薯条、冰咖啡、苏打水、甜点、小吃和可爱包装。柔和金色灯光，鲜艳温暖色调，光泽美食质感，奶油虚化效果，俏皮剪贴簿构图。每个物体周围有白色涂鸦轮廓，点缀手绘星星、心形、闪光、笑脸、云朵、箭头、小型卡通美食吉祥物。手写字体引言散布在图像各处，Instagram 美学，温馨“小事大幸福”氛围，韩日风格可爱编辑风格，高对比度，柔和阴影，梦幻氛围。画面混合逼真美食细节与异想天开的涂鸦艺术，贴纸式剪切效果，美学日记布局，潮流 Z 世代咖啡馆摄影，超详细，可爱心情板构图，电影感温暖色彩分级，怀旧舒适核心美学，4K 高质量。避免食物变形、文字乱码、过度杂乱、低清晰度、廉价贴纸感、脏污色彩、塑料食物质感。',
    ratio: '4:5',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['美食', '卡哇伊', '拼贴'],
    palette: ['#fef3c7', '#fb7185', '#22c55e'],
    visual: 'product'
  },
  {
    id: 'graduation-1980s-class',
    categoryId: 'graduation',
    title: '80 年代高中毕业照',
    mode: 'text_to_image',
    prompt: '80 年代中国高中毕业合影，一群高中生在校园教学楼前整齐排列，前排坐姿后排站立，朴素白衬衫、蓝色外套和学生装，神情认真克制，老式红砖教学楼背景，纪实摄影，轻微胶片颗粒，泛黄老照片质感。',
    ratio: '4:3',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['毕业照', '80年代', '纪实'],
    palette: ['#f5e6c8', '#64748b', '#7f1d1d'],
    visual: 'group'
  },
  {
    id: 'graduation-id-photo',
    categoryId: 'graduation',
    title: '清爽证件照',
    mode: 'image_to_image',
    prompt: '基于参考人物生成清爽证件照，保持身份特征一致，正面视角，肩部以上构图，白色或浅蓝背景，干净均匀布光，发型整洁，肤色自然，真实摄影，不夸张修饰。',
    ratio: '3:4',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '证件照', '正面'],
    palette: ['#dbeafe', '#ffffff', '#1e3a8a'],
    visual: 'portrait'
  },
  {
    id: 'graduation-university-lawn',
    categoryId: 'graduation',
    title: '大学草坪毕业合影',
    mode: 'text_to_image',
    prompt: '大学毕业季草坪合影，学生穿学士服抛帽，背景是校园主楼和绿树，阳光明亮，人物表情真实开心，构图开阔，青春纪念照风格，高清自然摄影。',
    ratio: '16:9',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['校园', '学士服', '合影'],
    palette: ['#dcfce7', '#111827', '#fbbf24'],
    visual: 'group'
  },
  {
    id: 'graduation-studio-formal',
    categoryId: 'graduation',
    title: '正式集体纪念照',
    mode: 'text_to_image',
    prompt: '正式集体纪念照，二十人左右在室内摄影棚整齐站坐排列，深色正装，背景为中性灰布景，均匀柔和灯光，构图端正，对称稳定，真实高清摄影。',
    ratio: '3:2',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['集体照', '正式', '棚拍'],
    palette: ['#f1f5f9', '#334155', '#0f172a'],
    visual: 'group'
  },


  
  {
    id: 'architecture-city-rain',
    categoryId: 'architecture',
    title: '雨夜城市街景',
    mode: 'text_to_image',
    prompt: '雨夜城市街景，高层建筑灯光倒映在湿润路面，行人撑伞经过，车灯形成柔和光轨，现代都市纪实摄影，冷暖对比，画面真实。',
    ratio: '21:9',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['城市', '雨夜', '纪实'],
    palette: ['#020617', '#2563eb', '#f97316'],
    visual: 'architecture'
  },
  {
    id: 'architecture-old-town',
    categoryId: 'architecture',
    title: '老城街巷晨光',
    mode: 'text_to_image',
    prompt: '清晨老城窄巷，青石板路、斑驳墙面、木窗和小店招牌，阳光从巷口照入，少量生活化人物，温暖纪实摄影，细节丰富。',
    ratio: '4:5',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['老城', '街巷', '晨光'],
    palette: ['#fef3c7', '#78716c', '#78350f'],
    visual: 'architecture'
  },
 
  {
    id: 'illustration-forest-book',
    categoryId: 'illustration',
    title: '森林绘本场景',
    mode: 'text_to_image',
    prompt: '儿童绘本插画，一条小路穿过温暖森林，小屋窗户亮着灯，树叶柔软蓬松，小动物藏在草丛边，温暖治愈，手绘水彩质感，细节丰富。',
    ratio: '4:5',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['绘本', '森林', '水彩'],
    palette: ['#dcfce7', '#fef3c7', '#16a34a'],
    visual: 'illustration'
  },
 
  {
    id: 'anime-avatar',
    categoryId: 'anime',
    title: '动漫头像转绘',
    mode: 'image_to_image',
    prompt: '基于参考人物生成二次元头像，保持发型、五官气质和主要身份特征，清爽动漫上色，明亮眼神，简洁浅色背景，适合社交头像。',
    ratio: '1:1',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '头像', '转绘'],
    palette: ['#ede9fe', '#a78bfa', '#0f172a'],
    visual: 'anime'
  },
 

  {
    id: 'anime-avatar-na-vi-selfie',
    categoryId: 'anime',
    title: '动漫角色合影',
    mode: 'image_to_image',
    prompt: '参考上传图片中的人物形象，保持人物的面部五官、肤色、发型、表情气质、身材比例和身份辨识度一致，生成一张人物在电影院里的逼真合影自拍照。参考图人物正与身旁的“阿凡达”电影中的纳威人（Na’vi）一起自拍，人物身穿一件印有“AVATAR”字样的黑色 T 恤和蓝色牛仔裤。背景是电影院座位和一个巨大的电影屏幕，屏幕上正在播放“阿凡达3：火与烬”的影片画面。整体画面真实可信，影院环境光自然，自拍视角亲近，人物和纳威人都清晰入镜，表情自然互动，细节丰富，高分辨率，8K 质感。避免人物脸部不一致、五官变形、身材比例改变、文字乱码、T 恤文字错误、背景杂乱、卡通化、塑料感、AI 痕迹。',
    ratio: '16:9',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '合影', '电影角色'],
    palette: ['#dbeafe', '#2563eb', '#0f172a'],
    visual: 'anime'
  },
  {
    id: 'anime-zootopia-cinema-selfie',
    categoryId: 'anime',
    title: '疯狂动物城合影',
    mode: 'image_to_image',
    prompt: '参考用户上传的照片中的人物角色，完整还原上传照片中的人物外貌、发型、妆容、服装、表情和身材比例，生成一张昏暗电影院里散场后的自拍合影。巨大银幕亮着蓝色中文大字“疯狂动物城2”。用户上传的人物居中，左侧紧贴着拟人化赤狐尼克·王尔德（Nick Wilde），真实细腻毛发，穿着绿色夏威夷花衬衫和紫色领带，笑着看向镜头；右侧紧贴着拟人化灰兔朱迪·霍普斯（Judy Hopps），真实细腻毛发，身穿蓝色警察制服，笑着看向镜头。三个人亲密靠在一起自拍合影，尼克和朱迪像好朋友一样贴近肩膀，电影院氛围浓郁，银幕蓝光作为主光源，周围氛围灯昏暗，摄影级真实感，毛发极致细节，超高清 8K 质感，锐利，专业电影院自拍，杰作，顶级画质。避免用户人物脸部不一致、五官变形、身材比例改变、角色毛发粗糙、角色服装错误、银幕文字乱码、背景杂乱、塑料感、卡通贴纸感、AI 痕迹。',
    ratio: '16:9',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['图生图', '疯狂动物城', '合影'],
    palette: ['#dbeafe', '#f97316', '#1e3a8a'],
    visual: 'anime'
  },
  {
    id: 'anime-3d-historical-biography-poster',
    categoryId: 'anime',
    title: '3D历史人物',
    mode: 'text_to_image',
    prompt: '根据用户输入的角色名字创作一张高完成度的「历史人物破框立像档案 / 3D Historical Breakout Biography Poster」。角色名字：苏轼。这不是普通历史海报，而是一张以「中央全身 3D 历史人物 + 破框前冲动作 + 四周轻质悬浮历史事件簇 + 底部轻量年谱时间轴」为核心结构的高完成度历史人物纪传海报。画面中央必须是一位完整全身、站立姿态的 3D 历史人物，人物是整张图最重要的主视觉，具有强体积感、雕塑感、纪念碑感和历史气场。服饰、冠帽、纹样、腰带、鞋履、配饰都要符合其时代身份，整体写实、厚重、有电影级冲击力。人物脚下有一个轻而稳的纪念碑式台基，负责承托人物，但不能厚重压画面。人物不能只是普通站立，必须有明显的前冲破框动作。手部、手臂或所持关键道具，如书卷、毛笔、卷轴、笏板、文稿等，必须朝向观者伸出，形成强烈近大远小、前后纵深、透视压缩和破出画面的 3D 冲击。前冲出来的关键道具必须完整可见、清晰可辨识，不能被画面裁切。围绕人物四周布置 6 到 10 个关键历史节点，不要做成厚重规整的小卡片，而要做成轻质悬浮、多朝向、半透明的历史事件簇。每个节点可由 2 到 4 个子碎片组成，包括编号、年份、事件标题、简短说明、小型历史场景画面、细金线、开放式边框、节点圆点、注释线等，像历史档案被拆解成轻盈的时间切片，在人物四周悬浮展开。事件簇有不同朝向和空间透视关系，有的左倾，有的右倾，有的向内转，有的向外翻，有的高，有的低，有前后层次，也可以被人物手臂、衣摆或道具轻微遮挡，形成空间纵深。它们具有玻璃感、亚克力感、微发光边缘、不完整边框和轻度错位叠层，信息详细但视觉上轻，不可压过中央人物。画面左侧设置大标题【苏轼】，使用具有东方历史气质的书法感、碑刻感或墨迹感字体，旁边可加入生卒年、身份标签、关键词总结和一句高度概括性的副标题，标题区稳、简、克制。底部设置一条轻盈、简洁、细致的横向时间轴，串联苏轼一生的重要年份，并与四周历史事件簇形成对应关系，像策展导览线、博物馆时间标尺或历史索引线，使用细金线、小圆点、简洁年份、小型文字标注和少量装饰节点，不要做成厚重黑色横栏或巨大底部信息块。背景以米白宣纸、古纸、旧档案纸或浅色博物馆墙面为基础，整体低饱和、干净、克制，可加入淡淡地图纹理、古建筑剪影、文献残片、印章、金色连接线、注释线、节点、疆域轮廓等元素，营造轻盈理性的历史档案展陈感。整张图必须建立明确对比：中央人物厚重、真实、立体、强光影、强体积；四周历史节点轻薄、透明、悬浮、微发光、精致理性；人物动作前冲破框，具有空间爆发力；底部时间轴轻盈克制，只负责收束信息。优先保证中央人物完整全身、前冲道具完整可见且有强烈 3D 冲击、四周事件簇轻盈悬浮不是厚重卡片、中央人物重而信息系统轻、底部时间轴轻量清晰不压画面。如果模型难以精确控制大量小字，请优先保证人物姓名、年份、编号、事件标题清晰，小段说明可适度简化或预留后期排版空间。',
    ratio: '4:5',
    resolution: '4K',
    quality: 'high',
    count: 1,
    tags: ['历史人物', '3D海报', '苏轼'],
    palette: ['#f8fafc', '#b45309', '#7c2d12'],
    visual: 'anime'
  },
 
]

const gallerySeedByID = new Map<string, GallerySeed>(seeds.map((item) => [item.id, item]))

export const imageStudioPromptGalleryItems: ImageStudioPromptGalleryItem[] = seeds.map((item) => ({
  id: item.id,
  categoryId: item.categoryId,
  title: item.title,
  mode: item.mode,
  prompt: item.prompt,
  ratio: item.ratio,
  resolution: item.resolution,
  quality: item.quality,
  count: item.count,
  image: item.image || `/image-studio/gallery/${encodeURIComponent(item.title)}.webp`,
  tags: item.tags
}))

export function imageStudioPromptGalleryFallback(item: Pick<ImageStudioPromptGalleryItem, 'id'>): string {
  const seed = gallerySeedByID.get(item.id)
  return seed ? makeGalleryPreview(seed) : ''
}

function makeGalleryPreview(item: GallerySeed): string {
  const category = imageStudioPromptGalleryCategories.find((entry) => entry.id === item.categoryId)
  const [background, accent, ink] = item.palette
  const visual = galleryVisual(item.visual, accent, ink)
  const label = item.mode === 'image_to_image' ? 'IMG2IMG' : 'TXT2IMG'
  const svg = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 960 720" role="img" aria-label="${escapeXml(item.title)}">
  <defs>
    <linearGradient id="bg-${item.id}" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="${background}"/>
      <stop offset="0.58" stop-color="#ffffff"/>
      <stop offset="1" stop-color="${accent}" stop-opacity="0.55"/>
    </linearGradient>
    <filter id="shadow-${item.id}" x="-20%" y="-20%" width="140%" height="140%">
      <feDropShadow dx="0" dy="18" stdDeviation="18" flood-color="#0f172a" flood-opacity="0.18"/>
    </filter>
  </defs>
  <rect width="960" height="720" fill="url(#bg-${item.id})"/>
  <circle cx="132" cy="104" r="72" fill="${accent}" opacity="0.18"/>
  <circle cx="794" cy="154" r="112" fill="${ink}" opacity="0.10"/>
  <path d="M0 596 C 174 520, 286 650, 456 584 S 762 512, 960 590 L960 720 L0 720 Z" fill="#ffffff" opacity="0.55"/>
  ${visual}
  <g filter="url(#shadow-${item.id})">
    <rect x="54" y="520" width="852" height="132" rx="24" fill="#ffffff" opacity="0.92"/>
    <text x="86" y="574" fill="#0f172a" font-family="Inter, Arial, sans-serif" font-size="38" font-weight="800">${escapeXml(item.title)}</text>
    <text x="86" y="618" fill="#475569" font-family="Inter, Arial, sans-serif" font-size="22">${escapeXml(category?.name || '')} · ${escapeXml(item.ratio)} · ${escapeXml(item.resolution)} · ${label}</text>
  </g>
</svg>`
  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`
}

function galleryVisual(visual: string, accent: string, ink: string): string {
  const common = `fill="${accent}" stroke="${ink}" stroke-width="12" stroke-linecap="round" stroke-linejoin="round"`
  if (visual === 'portrait') {
    return `<g transform="translate(354 128)">
      <circle cx="126" cy="116" r="82" fill="#fff" opacity="0.92"/>
      <path d="M56 306c18-86 120-120 204-66 38 24 60 64 66 112H30c3-16 12-32 26-46z" ${common} opacity="0.72"/>
      <circle cx="126" cy="102" r="70" ${common} opacity="0.76"/>
      <path d="M78 96c30-58 112-64 150-10" fill="none" stroke="${ink}" stroke-width="16"/>
    </g>`
  }
  if (visual === 'group') {
    return `<g transform="translate(230 164)" opacity="0.78">
      <rect x="88" y="208" width="324" height="112" rx="28" fill="#fff"/>
      <circle cx="150" cy="90" r="54" ${common}/>
      <circle cx="250" cy="72" r="58" ${common}/>
      <circle cx="350" cy="96" r="54" ${common}/>
      <path d="M70 318c18-74 94-108 150-62 12-54 110-82 160 2 56-36 116-8 142 60" fill="${accent}" stroke="${ink}" stroke-width="12"/>
    </g>`
  }
  if (visual === 'product') {
    return `<g transform="translate(342 126)" opacity="0.82">
      <ellipse cx="138" cy="384" rx="210" ry="34" fill="${ink}" opacity="0.15"/>
      <rect x="84" y="88" width="126" height="270" rx="34" ${common}/>
      <rect x="112" y="40" width="70" height="66" rx="16" fill="#fff" stroke="${ink}" stroke-width="12"/>
      <path d="M44 246h300M40 314h304" fill="none" stroke="#fff" stroke-width="18" opacity="0.76"/>
    </g>`
  }
  if (visual === 'interior') {
    return `<g transform="translate(158 142)" opacity="0.8">
      <rect x="70" y="86" width="500" height="292" rx="26" fill="#fff" stroke="${ink}" stroke-width="12"/>
      <rect x="122" y="130" width="148" height="114" rx="18" fill="${accent}" opacity="0.72"/>
      <path d="M318 252h182v126H120V278c48-42 128-44 198-26z" ${common}/>
      <path d="M166 426h328M210 378v48m236-48v48" stroke="${ink}" stroke-width="14" fill="none"/>
    </g>`
  }
  if (visual === 'architecture') {
    return `<g transform="translate(138 124)" opacity="0.82">
      <path d="M90 404h620M148 404V190l180-88 202 116 124-58v244" fill="#fff" stroke="${ink}" stroke-width="12"/>
      <path d="M148 190h506M270 132v272M430 176v228M574 196v208" fill="none" stroke="${accent}" stroke-width="22" opacity="0.72"/>
      <path d="M90 404c110-62 206-62 288 0 102-76 218-76 332 0" fill="none" stroke="${ink}" stroke-width="16" opacity="0.32"/>
    </g>`
  }
  if (visual === 'poster') {
    return `<g transform="translate(286 98)" opacity="0.82">
      <rect x="74" y="0" width="300" height="420" rx="26" fill="#fff" stroke="${ink}" stroke-width="12"/>
      <circle cx="224" cy="138" r="86" fill="${accent}" opacity="0.82"/>
      <path d="M126 288h196M126 332h140" stroke="${ink}" stroke-width="22" stroke-linecap="round"/>
      <path d="M96 78l256 238" stroke="${accent}" stroke-width="18" opacity="0.62"/>
    </g>`
  }
  if (visual === 'illustration') {
    return `<g transform="translate(150 150)" opacity="0.82">
      <path d="M110 342c30-132 166-222 286-148 34-82 162-90 210-10 76 8 122 58 130 138H110z" fill="#fff" stroke="${ink}" stroke-width="12"/>
      <circle cx="228" cy="132" r="54" fill="${accent}" stroke="${ink}" stroke-width="12"/>
      <path d="M96 352c128-80 256-78 384 0 70-48 150-50 238 0" fill="none" stroke="${accent}" stroke-width="22"/>
    </g>`
  }
  if (visual === 'chinese') {
    return `<g transform="translate(126 144)" opacity="0.84">
      <path d="M70 294c98-120 162-160 260-64 76-130 180-160 300 62" fill="none" stroke="${ink}" stroke-width="30" opacity="0.56"/>
      <path d="M90 368c132-52 250-56 354-12 86 36 168 28 246-20" fill="none" stroke="${accent}" stroke-width="18"/>
      <circle cx="650" cy="84" r="46" fill="${accent}" opacity="0.78"/>
      <path d="M220 120c-48 86-36 172 38 258" fill="none" stroke="${ink}" stroke-width="12"/>
    </g>`
  }
  if (visual === 'anime') {
    return `<g transform="translate(300 104)" opacity="0.84">
      <path d="M182 22l70 86 104 10-78 70 24 104-120-54-120 54 24-104-78-70 104-10z" fill="${accent}" stroke="${ink}" stroke-width="12"/>
      <circle cx="182" cy="214" r="86" fill="#fff" stroke="${ink}" stroke-width="12"/>
      <path d="M120 210c32 34 92 34 124 0M146 184h.1M220 184h.1" stroke="${ink}" stroke-width="18" stroke-linecap="round"/>
      <path d="M78 376c28-86 178-110 244 0" ${common}/>
    </g>`
  }
  return `<g transform="translate(126 150)" opacity="0.82">
    <rect x="86" y="56" width="540" height="342" rx="28" fill="#fff" stroke="${ink}" stroke-width="12"/>
    <path d="M120 260c118-82 242-92 372-30 54 26 100 28 168 4" fill="none" stroke="${accent}" stroke-width="24"/>
    <circle cx="220" cy="160" r="58" fill="${accent}" opacity="0.72"/>
    <path d="M156 344h388" stroke="${ink}" stroke-width="18" stroke-linecap="round"/>
  </g>`
}

function escapeXml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}
