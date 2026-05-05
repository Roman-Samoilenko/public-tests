-- -------------------------------------------------------
-- Миграция: официальные тесты по умолчанию
-- -------------------------------------------------------

-- Добавляем колонки если их ещё нет (из 000002)
ALTER TABLE tests ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE tests ADD COLUMN IF NOT EXISTS result_config JSONB;
ALTER TABLE test_answers ADD COLUMN IF NOT EXISTS result JSONB;
ALTER TABLE test_answers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE test_comments ADD COLUMN IF NOT EXISTS nickname VARCHAR(100) NOT NULL DEFAULT '';

-- -------------------------------------------------------
-- Тест 1: Политический компас
-- -------------------------------------------------------
INSERT INTO tests (author_id, title, description, questions, tags, result_config, status, is_official)
VALUES (
  0,
  'Политический компас',
  'Классический тест на определение политических взглядов по четырём осям: экономика, дипломатия, государство и общество. Отвечайте честно — здесь нет правильных или неправильных ответов.',
  $QUESTIONS$[
    {"id":"q1","text":"Угнетение со стороны корпораций вызывает больше опасений, чем угнетение со стороны государства.","type":"single_choice","required":true,"options":[{"id":"q1a","text":"Полностью согласен","scoring":{"econ":10,"govt":-5}},{"id":"q1b","text":"Скорее согласен","scoring":{"econ":5,"govt":-2}},{"id":"q1c","text":"Нейтрально","scoring":{}},{"id":"q1d","text":"Скорее не согласен","scoring":{"econ":-5,"govt":2}},{"id":"q1e","text":"Полностью не согласен","scoring":{"econ":-10,"govt":5}}]},
    {"id":"q2","text":"Государству необходимо вмешиваться в экономику для защиты потребителей.","type":"single_choice","required":true,"options":[{"id":"q2a","text":"Полностью согласен","scoring":{"econ":10}},{"id":"q2b","text":"Скорее согласен","scoring":{"econ":5}},{"id":"q2c","text":"Нейтрально","scoring":{}},{"id":"q2d","text":"Скорее не согласен","scoring":{"econ":-5}},{"id":"q2e","text":"Полностью не согласен","scoring":{"econ":-10}}]},
    {"id":"q3","text":"Чем свободнее рынки, тем свободнее люди.","type":"single_choice","required":true,"options":[{"id":"q3a","text":"Полностью согласен","scoring":{"econ":-10}},{"id":"q3b","text":"Скорее согласен","scoring":{"econ":-5}},{"id":"q3c","text":"Нейтрально","scoring":{}},{"id":"q3d","text":"Скорее не согласен","scoring":{"econ":5}},{"id":"q3e","text":"Полностью не согласен","scoring":{"econ":10}}]},
    {"id":"q4","text":"Лучше поддерживать сбалансированный бюджет, чем обеспечивать всеобщее благосостояние.","type":"single_choice","required":true,"options":[{"id":"q4a","text":"Полностью согласен","scoring":{"econ":-10}},{"id":"q4b","text":"Скорее согласен","scoring":{"econ":-5}},{"id":"q4c","text":"Нейтрально","scoring":{}},{"id":"q4d","text":"Скорее не согласен","scoring":{"econ":5}},{"id":"q4e","text":"Полностью не согласен","scoring":{"econ":10}}]},
    {"id":"q5","text":"Финансируемые государством исследования приносят людям больше пользы, чем рыночные.","type":"single_choice","required":true,"options":[{"id":"q5a","text":"Полностью согласен","scoring":{"econ":10,"scty":10}},{"id":"q5b","text":"Скорее согласен","scoring":{"econ":5,"scty":5}},{"id":"q5c","text":"Нейтрально","scoring":{}},{"id":"q5d","text":"Скорее не согласен","scoring":{"econ":-5,"scty":-5}},{"id":"q5e","text":"Полностью не согласен","scoring":{"econ":-10,"scty":-10}}]},
    {"id":"q6","text":"Таможенные пошлины важны для стимулирования отечественного производства.","type":"single_choice","required":true,"options":[{"id":"q6a","text":"Полностью согласен","scoring":{"econ":5,"govt":-10}},{"id":"q6b","text":"Скорее согласен","scoring":{"econ":2,"govt":-5}},{"id":"q6c","text":"Нейтрально","scoring":{}},{"id":"q6d","text":"Скорее не согласен","scoring":{"econ":-2,"govt":5}},{"id":"q6e","text":"Полностью не согласен","scoring":{"econ":-5,"govt":10}}]},
    {"id":"q7","text":"«От каждого по способностям, каждому по потребностям».","type":"single_choice","required":true,"options":[{"id":"q7a","text":"Полностью согласен","scoring":{"econ":10}},{"id":"q7b","text":"Скорее согласен","scoring":{"econ":5}},{"id":"q7c","text":"Нейтрально","scoring":{}},{"id":"q7d","text":"Скорее не согласен","scoring":{"econ":-5}},{"id":"q7e","text":"Полностью не согласен","scoring":{"econ":-10}}]},
    {"id":"q8","text":"Было бы лучше отменить социальные программы в пользу частной благотворительности.","type":"single_choice","required":true,"options":[{"id":"q8a","text":"Полностью согласен","scoring":{"econ":-10}},{"id":"q8b","text":"Скорее согласен","scoring":{"econ":-5}},{"id":"q8c","text":"Нейтрально","scoring":{}},{"id":"q8d","text":"Скорее не согласен","scoring":{"econ":5}},{"id":"q8e","text":"Полностью не согласен","scoring":{"econ":10}}]},
    {"id":"q9","text":"Налоги на богатых следует повысить, чтобы помочь бедным.","type":"single_choice","required":true,"options":[{"id":"q9a","text":"Полностью согласен","scoring":{"econ":10}},{"id":"q9b","text":"Скорее согласен","scoring":{"econ":5}},{"id":"q9c","text":"Нейтрально","scoring":{}},{"id":"q9d","text":"Скорее не согласен","scoring":{"econ":-5}},{"id":"q9e","text":"Полностью не согласен","scoring":{"econ":-10}}]},
    {"id":"q10","text":"Наследство — законная форма богатства.","type":"single_choice","required":true,"options":[{"id":"q10a","text":"Полностью согласен","scoring":{"econ":-10,"scty":-5}},{"id":"q10b","text":"Скорее согласен","scoring":{"econ":-5,"scty":-2}},{"id":"q10c","text":"Нейтрально","scoring":{}},{"id":"q10d","text":"Скорее не согласен","scoring":{"econ":5,"scty":2}},{"id":"q10e","text":"Полностью не согласен","scoring":{"econ":10,"scty":5}}]},
    {"id":"q11","text":"Базовые коммунальные услуги (дороги, электроэнергия) должны находиться в государственной собственности.","type":"single_choice","required":true,"options":[{"id":"q11a","text":"Полностью согласен","scoring":{"econ":10}},{"id":"q11b","text":"Скорее согласен","scoring":{"econ":5}},{"id":"q11c","text":"Нейтрально","scoring":{}},{"id":"q11d","text":"Скорее не согласен","scoring":{"econ":-5}},{"id":"q11e","text":"Полностью не согласен","scoring":{"econ":-10}}]},
    {"id":"q12","text":"Государственное вмешательство — угроза экономике.","type":"single_choice","required":true,"options":[{"id":"q12a","text":"Полностью согласен","scoring":{"econ":-10}},{"id":"q12b","text":"Скорее согласен","scoring":{"econ":-5}},{"id":"q12c","text":"Нейтрально","scoring":{}},{"id":"q12d","text":"Скорее не согласен","scoring":{"econ":5}},{"id":"q12e","text":"Полностью не согласен","scoring":{"econ":10}}]},
    {"id":"q13","text":"Качественная медицина должна быть доступна тем, кто может за неё платить.","type":"single_choice","required":true,"options":[{"id":"q13a","text":"Полностью согласен","scoring":{"econ":-10}},{"id":"q13b","text":"Скорее согласен","scoring":{"econ":-5}},{"id":"q13c","text":"Нейтрально","scoring":{}},{"id":"q13d","text":"Скорее не согласен","scoring":{"econ":5}},{"id":"q13e","text":"Полностью не согласен","scoring":{"econ":10}}]},
    {"id":"q14","text":"Качественное образование — право каждого человека.","type":"single_choice","required":true,"options":[{"id":"q14a","text":"Полностью согласен","scoring":{"econ":10,"scty":5}},{"id":"q14b","text":"Скорее согласен","scoring":{"econ":5,"scty":2}},{"id":"q14c","text":"Нейтрально","scoring":{}},{"id":"q14d","text":"Скорее не согласен","scoring":{"econ":-5,"scty":-2}},{"id":"q14e","text":"Полностью не согласен","scoring":{"econ":-10,"scty":-5}}]},
    {"id":"q15","text":"Средства производства должны принадлежать работникам, которые ими пользуются.","type":"single_choice","required":true,"options":[{"id":"q15a","text":"Полностью согласен","scoring":{"econ":10}},{"id":"q15b","text":"Скорее согласен","scoring":{"econ":5}},{"id":"q15c","text":"Нейтрально","scoring":{}},{"id":"q15d","text":"Скорее не согласен","scoring":{"econ":-5}},{"id":"q15e","text":"Полностью не согласен","scoring":{"econ":-10}}]},
    {"id":"q16","text":"ООН следует упразднить.","type":"single_choice","required":true,"options":[{"id":"q16a","text":"Полностью согласен","scoring":{"dipl":-10,"govt":-5}},{"id":"q16b","text":"Скорее согласен","scoring":{"dipl":-5,"govt":-2}},{"id":"q16c","text":"Нейтрально","scoring":{}},{"id":"q16d","text":"Скорее не согласен","scoring":{"dipl":5,"govt":2}},{"id":"q16e","text":"Полностью не согласен","scoring":{"dipl":10,"govt":5}}]},
    {"id":"q17","text":"Военные действия нашей страны часто необходимы для её защиты.","type":"single_choice","required":true,"options":[{"id":"q17a","text":"Полностью согласен","scoring":{"dipl":-10,"govt":-10}},{"id":"q17b","text":"Скорее согласен","scoring":{"dipl":-5,"govt":-5}},{"id":"q17c","text":"Нейтрально","scoring":{}},{"id":"q17d","text":"Скорее не согласен","scoring":{"dipl":5,"govt":5}},{"id":"q17e","text":"Полностью не согласен","scoring":{"dipl":10,"govt":10}}]},
    {"id":"q18","text":"Я поддерживаю региональные союзы, такие как Европейский союз.","type":"single_choice","required":true,"options":[{"id":"q18a","text":"Полностью согласен","scoring":{"econ":-5,"dipl":10,"govt":10,"scty":5}},{"id":"q18b","text":"Скорее согласен","scoring":{"econ":-2,"dipl":5,"govt":5,"scty":2}},{"id":"q18c","text":"Нейтрально","scoring":{}},{"id":"q18d","text":"Скорее не согласен","scoring":{"econ":2,"dipl":-5,"govt":-5,"scty":-2}},{"id":"q18e","text":"Полностью не согласен","scoring":{"econ":5,"dipl":-10,"govt":-10,"scty":-5}}]},
    {"id":"q19","text":"Важно сохранять национальный суверенитет.","type":"single_choice","required":true,"options":[{"id":"q19a","text":"Полностью согласен","scoring":{"dipl":-10,"govt":-5}},{"id":"q19b","text":"Скорее согласен","scoring":{"dipl":-5,"govt":-2}},{"id":"q19c","text":"Нейтрально","scoring":{}},{"id":"q19d","text":"Скорее не согласен","scoring":{"dipl":5,"govt":2}},{"id":"q19e","text":"Полностью не согласен","scoring":{"dipl":10,"govt":5}}]},
    {"id":"q20","text":"Единое мировое правительство пошло бы на пользу человечеству.","type":"single_choice","required":true,"options":[{"id":"q20a","text":"Полностью согласен","scoring":{"dipl":10}},{"id":"q20b","text":"Скорее согласен","scoring":{"dipl":5}},{"id":"q20c","text":"Нейтрально","scoring":{}},{"id":"q20d","text":"Скорее не согласен","scoring":{"dipl":-5}},{"id":"q20e","text":"Полностью не согласен","scoring":{"dipl":-10}}]},
    {"id":"q21","text":"Поддерживать мирные отношения важнее, чем наращивать мощь государства.","type":"single_choice","required":true,"options":[{"id":"q21a","text":"Полностью согласен","scoring":{"dipl":10}},{"id":"q21b","text":"Скорее согласен","scoring":{"dipl":5}},{"id":"q21c","text":"Нейтрально","scoring":{}},{"id":"q21d","text":"Скорее не согласен","scoring":{"dipl":-5}},{"id":"q21e","text":"Полностью не согласен","scoring":{"dipl":-10}}]},
    {"id":"q22","text":"Войны не нуждаются в оправдании перед другими странами.","type":"single_choice","required":true,"options":[{"id":"q22a","text":"Полностью согласен","scoring":{"dipl":-10,"govt":-10}},{"id":"q22b","text":"Скорее согласен","scoring":{"dipl":-5,"govt":-5}},{"id":"q22c","text":"Нейтрально","scoring":{}},{"id":"q22d","text":"Скорее не согласен","scoring":{"dipl":5,"govt":5}},{"id":"q22e","text":"Полностью не согласен","scoring":{"dipl":10,"govt":10}}]},
    {"id":"q23","text":"Военные расходы — это зачастую необоснованная трата денег.","type":"single_choice","required":true,"options":[{"id":"q23a","text":"Полностью согласен","scoring":{"dipl":10,"govt":10}},{"id":"q23b","text":"Скорее согласен","scoring":{"dipl":5,"govt":5}},{"id":"q23c","text":"Нейтрально","scoring":{}},{"id":"q23d","text":"Скорее не согласен","scoring":{"dipl":-5,"govt":-5}},{"id":"q23e","text":"Полностью не согласен","scoring":{"dipl":-10,"govt":-10}}]},
    {"id":"q24","text":"Международная помощь — это зачастую необоснованная трата денег.","type":"single_choice","required":true,"options":[{"id":"q24a","text":"Полностью согласен","scoring":{"econ":-5,"dipl":-10}},{"id":"q24b","text":"Скорее согласен","scoring":{"econ":-2,"dipl":-5}},{"id":"q24c","text":"Нейтрально","scoring":{}},{"id":"q24d","text":"Скорее не согласен","scoring":{"econ":2,"dipl":5}},{"id":"q24e","text":"Полностью не согласен","scoring":{"econ":5,"dipl":10}}]},
    {"id":"q25","text":"Исследования должны проводиться в международном масштабе.","type":"single_choice","required":true,"options":[{"id":"q25a","text":"Полностью согласен","scoring":{"dipl":10,"scty":10}},{"id":"q25b","text":"Скорее согласен","scoring":{"dipl":5,"scty":5}},{"id":"q25c","text":"Нейтрально","scoring":{}},{"id":"q25d","text":"Скорее не согласен","scoring":{"dipl":-5,"scty":-5}},{"id":"q25e","text":"Полностью не согласен","scoring":{"dipl":-10,"scty":-10}}]},
    {"id":"q26","text":"Правительства должны быть подотчётны международному сообществу.","type":"single_choice","required":true,"options":[{"id":"q26a","text":"Полностью согласен","scoring":{"dipl":10,"govt":5}},{"id":"q26b","text":"Скорее согласен","scoring":{"dipl":5,"govt":2}},{"id":"q26c","text":"Нейтрально","scoring":{}},{"id":"q26d","text":"Скорее не согласен","scoring":{"dipl":-5,"govt":-2}},{"id":"q26e","text":"Полностью не согласен","scoring":{"dipl":-10,"govt":-5}}]},
    {"id":"q27","text":"Общество лучше всего управляется при сильном централизованном руководстве.","type":"single_choice","required":true,"options":[{"id":"q27a","text":"Полностью согласен","scoring":{"govt":-10}},{"id":"q27b","text":"Скорее согласен","scoring":{"govt":-5}},{"id":"q27c","text":"Нейтрально","scoring":{}},{"id":"q27d","text":"Скорее не согласен","scoring":{"govt":5}},{"id":"q27e","text":"Полностью не согласен","scoring":{"govt":10}}]},
    {"id":"q28","text":"Эвтаназия и ассистированный суицид должны быть легальны.","type":"single_choice","required":true,"options":[{"id":"q28a","text":"Полностью согласен","scoring":{"govt":10}},{"id":"q28b","text":"Скорее согласен","scoring":{"govt":5}},{"id":"q28c","text":"Нейтрально","scoring":{}},{"id":"q28d","text":"Скорее не согласен","scoring":{"govt":-5}},{"id":"q28e","text":"Полностью не согласен","scoring":{"govt":-10}}]},
    {"id":"q29","text":"Ради защиты от терроризма допустимо пожертвовать частью гражданских свобод.","type":"single_choice","required":true,"options":[{"id":"q29a","text":"Полностью согласен","scoring":{"govt":-10}},{"id":"q29b","text":"Скорее согласен","scoring":{"govt":-5}},{"id":"q29c","text":"Нейтрально","scoring":{}},{"id":"q29d","text":"Скорее не согласен","scoring":{"govt":5}},{"id":"q29e","text":"Полностью не согласен","scoring":{"govt":10}}]},
    {"id":"q30","text":"Государственная слежка необходима в современном мире.","type":"single_choice","required":true,"options":[{"id":"q30a","text":"Полностью согласен","scoring":{"govt":-10}},{"id":"q30b","text":"Скорее согласен","scoring":{"govt":-5}},{"id":"q30c","text":"Нейтрально","scoring":{}},{"id":"q30d","text":"Скорее не согласен","scoring":{"govt":5}},{"id":"q30e","text":"Полностью не согласен","scoring":{"govt":10}}]},
    {"id":"q31","text":"Само существование государства — угроза нашей свободе.","type":"single_choice","required":true,"options":[{"id":"q31a","text":"Полностью согласен","scoring":{"govt":10}},{"id":"q31b","text":"Скорее согласен","scoring":{"govt":5}},{"id":"q31c","text":"Нейтрально","scoring":{}},{"id":"q31d","text":"Скорее не согласен","scoring":{"govt":-5}},{"id":"q31e","text":"Полностью не согласен","scoring":{"govt":-10}}]},
    {"id":"q32","text":"Любая власть заслуживает сомнения.","type":"single_choice","required":true,"options":[{"id":"q32a","text":"Полностью согласен","scoring":{"govt":10,"scty":5}},{"id":"q32b","text":"Скорее согласен","scoring":{"govt":5,"scty":2}},{"id":"q32c","text":"Нейтрально","scoring":{}},{"id":"q32d","text":"Скорее не согласен","scoring":{"govt":-5,"scty":-2}},{"id":"q32e","text":"Полностью не согласен","scoring":{"govt":-10,"scty":-5}}]},
    {"id":"q33","text":"Иерархическое государство — лучшая форма правления.","type":"single_choice","required":true,"options":[{"id":"q33a","text":"Полностью согласен","scoring":{"govt":-10}},{"id":"q33b","text":"Скорее согласен","scoring":{"govt":-5}},{"id":"q33c","text":"Нейтрально","scoring":{}},{"id":"q33d","text":"Скорее не согласен","scoring":{"govt":5}},{"id":"q33e","text":"Полностью не согласен","scoring":{"govt":10}}]},
    {"id":"q34","text":"Демократия — это нечто большее, чем просто процесс принятия решений.","type":"single_choice","required":true,"options":[{"id":"q34a","text":"Полностью согласен","scoring":{"govt":10}},{"id":"q34b","text":"Скорее согласен","scoring":{"govt":5}},{"id":"q34c","text":"Нейтрально","scoring":{}},{"id":"q34d","text":"Скорее не согласен","scoring":{"govt":-5}},{"id":"q34e","text":"Полностью не согласен","scoring":{"govt":-10}}]},
    {"id":"q35","text":"Экологические нормы необходимы.","type":"single_choice","required":true,"options":[{"id":"q35a","text":"Полностью согласен","scoring":{"econ":5,"scty":10}},{"id":"q35b","text":"Скорее согласен","scoring":{"econ":2,"scty":5}},{"id":"q35c","text":"Нейтрально","scoring":{}},{"id":"q35d","text":"Скорее не согласен","scoring":{"econ":-2,"scty":-5}},{"id":"q35e","text":"Полностью не согласен","scoring":{"econ":-5,"scty":-10}}]},
    {"id":"q36","text":"Лучший мир придёт через автоматизацию, науку и технологии.","type":"single_choice","required":true,"options":[{"id":"q36a","text":"Полностью согласен","scoring":{"scty":10}},{"id":"q36b","text":"Скорее согласен","scoring":{"scty":5}},{"id":"q36c","text":"Нейтрально","scoring":{}},{"id":"q36d","text":"Скорее не согласен","scoring":{"scty":-5}},{"id":"q36e","text":"Полностью не согласен","scoring":{"scty":-10}}]},
    {"id":"q37","text":"Детей нужно воспитывать в религиозных или традиционных ценностях.","type":"single_choice","required":true,"options":[{"id":"q37a","text":"Полностью согласен","scoring":{"govt":-5,"scty":-10}},{"id":"q37b","text":"Скорее согласен","scoring":{"govt":-2,"scty":-5}},{"id":"q37c","text":"Нейтрально","scoring":{}},{"id":"q37d","text":"Скорее не согласен","scoring":{"govt":2,"scty":5}},{"id":"q37e","text":"Полностью не согласен","scoring":{"govt":5,"scty":10}}]},
    {"id":"q38","text":"Традиции не имеют ценности сами по себе.","type":"single_choice","required":true,"options":[{"id":"q38a","text":"Полностью согласен","scoring":{"scty":10}},{"id":"q38b","text":"Скорее согласен","scoring":{"scty":5}},{"id":"q38c","text":"Нейтрально","scoring":{}},{"id":"q38d","text":"Скорее не согласен","scoring":{"scty":-5}},{"id":"q38e","text":"Полностью не согласен","scoring":{"scty":-10}}]},
    {"id":"q39","text":"Религия должна играть роль в государственном управлении.","type":"single_choice","required":true,"options":[{"id":"q39a","text":"Полностью согласен","scoring":{"govt":-10,"scty":-10}},{"id":"q39b","text":"Скорее согласен","scoring":{"govt":-5,"scty":-5}},{"id":"q39c","text":"Нейтрально","scoring":{}},{"id":"q39d","text":"Скорее не согласен","scoring":{"govt":5,"scty":5}},{"id":"q39e","text":"Полностью не согласен","scoring":{"govt":10,"scty":10}}]},
    {"id":"q40","text":"Церкви должны облагаться налогами так же, как другие организации.","type":"single_choice","required":true,"options":[{"id":"q40a","text":"Полностью согласен","scoring":{"econ":5,"scty":10}},{"id":"q40b","text":"Скорее согласен","scoring":{"econ":2,"scty":5}},{"id":"q40c","text":"Нейтрально","scoring":{}},{"id":"q40d","text":"Скорее не согласен","scoring":{"econ":-2,"scty":-5}},{"id":"q40e","text":"Полностью не согласен","scoring":{"econ":-5,"scty":-10}}]},
    {"id":"q41","text":"Изменение климата — одна из главных угроз нашей жизни.","type":"single_choice","required":true,"options":[{"id":"q41a","text":"Полностью согласен","scoring":{"scty":10}},{"id":"q41b","text":"Скорее согласен","scoring":{"scty":5}},{"id":"q41c","text":"Нейтрально","scoring":{}},{"id":"q41d","text":"Скорее не согласен","scoring":{"scty":-5}},{"id":"q41e","text":"Полностью не согласен","scoring":{"scty":-10}}]},
    {"id":"q42","text":"Важно объединиться всему миру в борьбе с изменением климата.","type":"single_choice","required":true,"options":[{"id":"q42a","text":"Полностью согласен","scoring":{"dipl":10,"scty":10}},{"id":"q42b","text":"Скорее согласен","scoring":{"dipl":5,"scty":5}},{"id":"q42c","text":"Нейтрально","scoring":{}},{"id":"q42d","text":"Скорее не согласен","scoring":{"dipl":-5,"scty":-5}},{"id":"q42e","text":"Полностью не согласен","scoring":{"dipl":-10,"scty":-10}}]},
    {"id":"q43","text":"Общество было лучше много лет назад, чем сейчас.","type":"single_choice","required":true,"options":[{"id":"q43a","text":"Полностью согласен","scoring":{"scty":-10}},{"id":"q43b","text":"Скорее согласен","scoring":{"scty":-5}},{"id":"q43c","text":"Нейтрально","scoring":{}},{"id":"q43d","text":"Скорее не согласен","scoring":{"scty":5}},{"id":"q43e","text":"Полностью не согласен","scoring":{"scty":10}}]},
    {"id":"q44","text":"Важно сохранять традиции прошлого.","type":"single_choice","required":true,"options":[{"id":"q44a","text":"Полностью согласен","scoring":{"scty":-10}},{"id":"q44b","text":"Скорее согласен","scoring":{"scty":-5}},{"id":"q44c","text":"Нейтрально","scoring":{}},{"id":"q44d","text":"Скорее не согласен","scoring":{"scty":5}},{"id":"q44e","text":"Полностью не согласен","scoring":{"scty":10}}]},
    {"id":"q45","text":"Важно думать о долгосрочной перспективе, выходящей за рамки нашей жизни.","type":"single_choice","required":true,"options":[{"id":"q45a","text":"Полностью согласен","scoring":{"scty":10}},{"id":"q45b","text":"Скорее согласен","scoring":{"scty":5}},{"id":"q45c","text":"Нейтрально","scoring":{}},{"id":"q45d","text":"Скорее не согласен","scoring":{"scty":-5}},{"id":"q45e","text":"Полностью не согласен","scoring":{"scty":-10}}]},
    {"id":"q46","text":"Использование наркотиков следует легализовать или декриминализировать.","type":"single_choice","required":true,"options":[{"id":"q46a","text":"Полностью согласен","scoring":{"govt":10,"scty":2}},{"id":"q46b","text":"Скорее согласен","scoring":{"govt":5,"scty":1}},{"id":"q46c","text":"Нейтрально","scoring":{}},{"id":"q46d","text":"Скорее не согласен","scoring":{"govt":-5,"scty":-1}},{"id":"q46e","text":"Полностью не согласен","scoring":{"govt":-10,"scty":-2}}]},
    {"id":"q47","text":"Однополые браки должны быть легальны.","type":"single_choice","required":true,"options":[{"id":"q47a","text":"Полностью согласен","scoring":{"govt":10,"scty":10}},{"id":"q47b","text":"Скорее согласен","scoring":{"govt":5,"scty":5}},{"id":"q47c","text":"Нейтрально","scoring":{}},{"id":"q47d","text":"Скорее не согласен","scoring":{"govt":-5,"scty":-5}},{"id":"q47e","text":"Полностью не согласен","scoring":{"govt":-10,"scty":-10}}]},
    {"id":"q48","text":"Ни одна культура не превосходит другие.","type":"single_choice","required":true,"options":[{"id":"q48a","text":"Полностью согласен","scoring":{"dipl":10,"govt":5,"scty":10}},{"id":"q48b","text":"Скорее согласен","scoring":{"dipl":5,"govt":2,"scty":5}},{"id":"q48c","text":"Нейтрально","scoring":{}},{"id":"q48d","text":"Скорее не согласен","scoring":{"dipl":-5,"govt":-2,"scty":-5}},{"id":"q48e","text":"Полностью не согласен","scoring":{"dipl":-10,"govt":-5,"scty":-10}}]},
    {"id":"q49","text":"Если мы принимаем мигрантов, им важно ассимилироваться в нашей культуре.","type":"single_choice","required":true,"options":[{"id":"q49a","text":"Полностью согласен","scoring":{"govt":-5,"scty":-10}},{"id":"q49b","text":"Скорее согласен","scoring":{"govt":-2,"scty":-5}},{"id":"q49c","text":"Нейтрально","scoring":{}},{"id":"q49d","text":"Скорее не согласен","scoring":{"govt":2,"scty":5}},{"id":"q49e","text":"Полностью не согласен","scoring":{"govt":5,"scty":10}}]},
    {"id":"q50","text":"Аборты должны быть запрещены в большинстве или всех случаях.","type":"single_choice","required":true,"options":[{"id":"q50a","text":"Полностью согласен","scoring":{"govt":-10,"scty":-10}},{"id":"q50b","text":"Скорее согласен","scoring":{"govt":-5,"scty":-5}},{"id":"q50c","text":"Нейтрально","scoring":{}},{"id":"q50d","text":"Скорее не согласен","scoring":{"govt":5,"scty":5}},{"id":"q50e","text":"Полностью не согласен","scoring":{"govt":10,"scty":10}}]},
    {"id":"q51","text":"Я поддерживаю всеобщее медицинское страхование, оплачиваемое государством.","type":"single_choice","required":true,"options":[{"id":"q51a","text":"Полностью согласен","scoring":{"econ":10}},{"id":"q51b","text":"Скорее согласен","scoring":{"econ":5}},{"id":"q51c","text":"Нейтрально","scoring":{}},{"id":"q51d","text":"Скорее не согласен","scoring":{"econ":-5}},{"id":"q51e","text":"Полностью не согласен","scoring":{"econ":-10}}]},
    {"id":"q52","text":"Мы должны открыть наши границы для иммиграции.","type":"single_choice","required":true,"options":[{"id":"q52a","text":"Полностью согласен","scoring":{"dipl":10,"govt":10}},{"id":"q52b","text":"Скорее согласен","scoring":{"dipl":5,"govt":5}},{"id":"q52c","text":"Нейтрально","scoring":{}},{"id":"q52d","text":"Скорее не согласен","scoring":{"dipl":-5,"govt":-5}},{"id":"q52e","text":"Полностью не согласен","scoring":{"dipl":-10,"govt":-10}}]},
    {"id":"q53","text":"Все люди — вне зависимости от культуры, сексуальности и других факторов — должны иметь равные права.","type":"single_choice","required":true,"options":[{"id":"q53a","text":"Полностью согласен","scoring":{"econ":10,"dipl":10,"govt":10,"scty":10}},{"id":"q53b","text":"Скорее согласен","scoring":{"econ":5,"dipl":5,"govt":5,"scty":5}},{"id":"q53c","text":"Нейтрально","scoring":{}},{"id":"q53d","text":"Скорее не согласен","scoring":{"econ":-5,"dipl":-5,"govt":-5,"scty":-5}},{"id":"q53e","text":"Полностью не согласен","scoring":{"econ":-10,"dipl":-10,"govt":-10,"scty":-10}}]}
  ]$QUESTIONS$::jsonb,
  ARRAY['политика','общество','ценности'],
  $RCONFIG${
    "type": "combo",
    "description": "Политический компас — это способ визуализировать политические взгляды по двум осям: экономической (левые–правые) и социальной (авторитаризм–либертарианство). Ваш результат отражает склонности, а не ярлык.",
    "axes": [
      {"id":"econ","label":"Экономика","min":-100,"max":100,"left_label":"Правые (рынок)","right_label":"Левые (государство)"},
      {"id":"dipl","label":"Дипломатия","min":-100,"max":100,"left_label":"Национализм","right_label":"Глобализм"},
      {"id":"govt","label":"Государство","min":-100,"max":100,"left_label":"Авторитаризм","right_label":"Либертарианство"},
      {"id":"scty","label":"Общество","min":-100,"max":100,"left_label":"Традиционализм","right_label":"Прогрессивизм"}
    ],
    "results": [
      {"key":"auth_left",    "label":"Авторитарные левые",     "description":"Вы выступаете за сильное государство, которое контролирует экономику и управляет обществом ради равенства. Социализм с централизованной властью.","target":{"econ":70,"dipl":0,"govt":-70,"scty":0}},
      {"key":"lib_left",     "label":"Либертарные левые",      "description":"Вы за экономическое равенство и личную свободу одновременно. Анархизм, либертарный социализм, экосоциализм.","target":{"econ":70,"dipl":20,"govt":70,"scty":70}},
      {"key":"auth_right",   "label":"Авторитарные правые",    "description":"Вы за свободный рынок, но в рамках сильного государства с традиционными ценностями. Национал-консерватизм.","target":{"econ":-70,"dipl":-50,"govt":-70,"scty":-70}},
      {"key":"lib_right",    "label":"Либертарные правые",     "description":"Вы за свободный рынок и минимальное государство. Классический либерализм, либертарианство.","target":{"econ":-70,"dipl":0,"govt":70,"scty":20}},
      {"key":"centrist",     "label":"Центрист",                "description":"Ваши взгляды сбалансированы по всем осям. Вы прагматик, не склонный к крайностям.","target":{"econ":0,"dipl":0,"govt":0,"scty":0}},
      {"key":"progressive",  "label":"Прогрессист",             "description":"Вы цените социальный прогресс, науку и реформы. Умеренно левые взгляды с либеральными социальными позициями.","target":{"econ":40,"dipl":40,"govt":40,"scty":80}},
      {"key":"conservative", "label":"Консерватор",             "description":"Вы цените стабильность, традиции и проверенные временем институты. Умеренно правые взгляды.","target":{"econ":-40,"dipl":-40,"govt":-30,"scty":-70}},
      {"key":"nationalist",  "label":"Националист",             "description":"Интересы своей страны для вас на первом месте. Скептицизм к международным институтам и открытым границам.","target":{"econ":0,"dipl":-80,"govt":-30,"scty":-30}}
    ]
  }$RCONFIG$::jsonb,
  'published',
  true
);

-- -------------------------------------------------------
-- Тест 2: Первое знакомство
-- -------------------------------------------------------
INSERT INTO tests (author_id, title, description, questions, tags, result_config, status, is_official)
VALUES (
  0,
  'Первое знакомство',
  'Расскажите немного о себе. Эти данные помогут платформе показывать более интересные результаты и корреляции. Ничего личного — только то, что вы хотите рассказать.',
  $QUESTIONS2$[
    {
      "id":"fq1",
      "text":"Сколько вам лет?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq1a","text":"До 18 лет"},
        {"id":"fq1b","text":"18–24 года"},
        {"id":"fq1c","text":"25–34 года"},
        {"id":"fq1d","text":"35–44 года"},
        {"id":"fq1e","text":"45–54 года"},
        {"id":"fq1f","text":"55 лет и старше"}
      ]
    },
    {
      "id":"fq2",
      "text":"Ваш пол",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq2a","text":"Мужской"},
        {"id":"fq2b","text":"Женский"},
        {"id":"fq2c","text":"Другой / предпочитаю не указывать"}
      ]
    },
    {
      "id":"fq3",
      "text":"Как вы себя ощущаете?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq3a","text":"Скорее технарь — логика, точные науки, системное мышление"},
        {"id":"fq3b","text":"Скорее гуманитарий — люди, слова, смыслы, творчество"},
        {"id":"fq3c","text":"Смешанный тип — одинаково комфортно с обеими сферами"},
        {"id":"fq3d","text":"Затрудняюсь ответить"}
      ]
    },
    {
      "id":"fq4",
      "text":"Вы сдавали ЕГЭ или аналогичный государственный экзамен для поступления в вуз? Если да, то какой у вас был средний балл?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq4a","text":"Да, 90+ баллов"},
        {"id":"fq4b","text":"Да, 80-90 баллов"},
        {"id":"fq4c","text":"Да, 60–79 баллов"},
        {"id":"fq4d","text":"Да, ниже 60 баллов"},
        {"id":"fq4e","text":"Не сдавал(а) ЕГЭ (другая страна, другое время или другой путь)"}
      ]
    },
    {
      "id":"fq5",
      "text":"Вы скорее интроверт или экстраверт?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq5a","text":"Интроверт — заряжаюсь в одиночестве, общение утомляет"},
        {"id":"fq5b","text":"Экстраверт — заряжаюсь от людей, тишина угнетает"},
        {"id":"fq5c","text":"Амбиверт — зависит от настроения и ситуации"},
        {"id":"fq5d","text":"Не думал(а) об этом / не знаю"}
      ]
    },
    {
      "id":"fq6",
      "text":"Насколько вы доверяете своей интуиции при принятии решений?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq6a","text":"Очень доверяю — часто действую по ощущению"},
        {"id":"fq6b","text":"Скорее доверяю, но перепроверяю"},
        {"id":"fq6c","text":"Скорее не доверяю — предпочитаю факты и логику"},
        {"id":"fq6d","text":"Совсем не доверяю интуиции"}
      ]
    },
    {
      "id":"fq7",
      "text":"Как вы относитесь к переменам в жизни?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq7a","text":"Люблю перемены — скука хуже неизвестности"},
        {"id":"fq7b","text":"Принимаю перемены, хотя они и вызывают тревогу"},
        {"id":"fq7c","text":"Предпочитаю стабильность, но адаптируюсь"},
        {"id":"fq7d","text":"Стабильность — моё всё, перемены стрессуют"}
      ]
    },
    {
      "id":"fq8",
      "text":"Как вы предпочитаете проводить свободное время?",
      "type":"multiple_choice",
      "required":false,
      "options":[
        {"id":"fq8a","text":"Чтение книг или статей"},
        {"id":"fq8b","text":"Видеоигры или сериалы"},
        {"id":"fq8c","text":"Спорт или активный отдых"},
        {"id":"fq8d","text":"Общение с друзьями или семьёй"},
        {"id":"fq8e","text":"Творчество (музыка, рисование, письмо)"},
        {"id":"fq8f","text":"Технологии, программирование, DIY"},
        {"id":"fq8g","text":"Путешествия и новые впечатления"}
      ]
    },
    {
      "id":"fq9",
      "text":"Какой у вас стиль работы?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq9a","text":"Делаю всё заранее — не люблю дедлайны"},
        {"id":"fq9b","text":"Планирую, но иногда откладываю"},
        {"id":"fq9c","text":"Лучше работаю в последний момент под давлением"},
        {"id":"fq9d","text":"Всё зависит от задачи и настроения"}
      ]
    },
    {
      "id":"fq10",
      "text":"Что для вас важнее в работе или учёбе?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq10a","text":"Результат — главное получить нужный итог"},
        {"id":"fq10b","text":"Процесс — важно получать удовольствие от самой работы"},
        {"id":"fq10c","text":"Признание — приятно, когда труд замечают"},
        {"id":"fq10d","text":"Развитие — хочу становиться лучше с каждым разом"}
      ]
    },
    {
      "id":"fq11",
      "text":"Как вы относитесь к спорам и дискуссиям?",
      "type":"single_choice",
      "required":false,
      "options":[
        {"id":"fq11a","text":"Люблю дискуссии — это способ найти истину"},
        {"id":"fq11b","text":"Участвую, если тема важна для меня"},
        {"id":"fq11c","text":"Стараюсь избегать конфликтов"},
        {"id":"fq11d","text":"Не люблю споры — предпочитаю согласие"}
      ]
    },
    {
      "id":"fq12",
      "text":"Как вы узнаёте новости и информацию о мире?",
      "type":"multiple_choice",
      "required":false,
      "options":[
        {"id":"fq12a","text":"Телеграм-каналы"},
        {"id":"fq12b","text":"Социальные сети (ВК, Instagram и др.)"},
        {"id":"fq12c","text":"Новостные сайты и СМИ"},
        {"id":"fq12d","text":"YouTube и подкасты"},
        {"id":"fq12e","text":"Из разговоров с людьми"},
        {"id":"fq12f","text":"Практически не слежу за новостями"}
      ]
    }
  ]$QUESTIONS2$::jsonb,
  ARRAY['знакомство','личность','профиль'],
  $RCONFIG2${"type":"none"}$RCONFIG2$::jsonb,
  'published',
  true
);

-- Логируем создание официальных тестов
INSERT INTO moderation_log (test_id, author_id, title, action)
SELECT id, author_id, title, 'created'
FROM tests
WHERE is_official = true
  AND created_at > NOW() - INTERVAL '5 seconds';
