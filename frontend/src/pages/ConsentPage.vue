<template>
  <div class="page legal-page">
    <router-link to="/" class="back-link font-mono">← Лента</router-link>
    <div class="legal-header">
      <span class="legal-label font-mono">ПРАВОВЫЕ ДОКУМЕНТЫ</span>
      <h1 class="legal-title">Согласие на обработку персональных данных</h1>
      <p class="legal-meta font-mono">Редакция от 01.01.2025</p>
    </div>

    <div class="divider"></div>

    <div class="legal-body">

      <div class="consent-notice">
        <span class="notice-icon font-mono">i</span>
        <p>
          Данный документ носит информационный характер. Принятие его условий происходит
          автоматически в момент регистрации на Платформе.
        </p>
      </div>

      <section class="legal-section">
        <h2>Субъект и оператор</h2>
        <p>
          Настоящее согласие даётся физическим лицом (Субъектом персональных данных, далее — «Субъект»)
          Администрации сайта «public-tests» (далее — «Оператор») в соответствии со статьёй 9
          Федерального закона от 27.07.2006 № 152-ФЗ «О персональных данных».
        </p>
      </section>

      <section class="legal-section">
        <h2>Перечень обрабатываемых данных</h2>
        <p>Субъект даёт согласие на обработку следующих персональных данных:</p>
        <ul>
          <li>Адрес электронной почты.</li>
          <li>Публичный никнейм (псевдоним).</li>
          <li>Ответы на опросы и тесты.</li>
          <li>Пользовательский контент (созданные тесты, комментарии).</li>
          <li>Технические идентификаторы сессии.</li>
        </ul>
      </section>

      <section class="legal-section">
        <h2>Цели обработки</h2>
        <ul>
          <li>Аутентификация и авторизация на Платформе.</li>
          <li>Обеспечение работы функций Платформы.</li>
          <li>Направление одноразовых кодов подтверждения (OTP).</li>
          <li>Предотвращение злоупотреблений и обеспечение безопасности.</li>
        </ul>
      </section>

      <section class="legal-section">
        <h2>Способы обработки</h2>
        <p>
          Персональные данные обрабатываются смешанным способом (с использованием средств
          автоматизации и без таковых): сбор, запись, систематизация, накопление, хранение,
          использование, обезличивание, удаление, уничтожение.
        </p>
      </section>

      <section class="legal-section">
        <h2>Передача третьим лицам</h2>
        <p>
          Персональные данные не передаются третьим лицам в коммерческих целях. Передача
          возможна исключительно по требованию уполномоченных государственных органов в
          порядке, установленном законодательством РФ.
        </p>
      </section>

      <section class="legal-section">
        <h2>Срок действия согласия</h2>
        <p>
          Согласие действует с момента регистрации до момента удаления аккаунта или отзыва
          согласия Субъектом. После отзыва согласия Оператор прекращает обработку персональных
          данных и уничтожает их в разумный срок, за исключением случаев, когда обязанность по
          хранению предусмотрена законодательством.
        </p>
      </section>

      <section class="legal-section">
        <h2>Порядок отзыва согласия</h2>
        <p>
          Субъект вправе в любой момент отозвать настоящее согласие, направив обращение
          на контактный адрес электронной почты Администрации. Отзыв согласия не влечёт
          незаконности обработки, осуществлённой до его получения.
        </p>
      </section>

    </div>
  </div>
</template>

<script setup></script>

<style scoped>
.back-link {
  display: inline-block; font-size: 0.78rem; letter-spacing: 0.06em;
  color: var(--text-muted); margin-bottom: 1.2rem; transition: color var(--transition);
}
.back-link:hover { color: var(--accent); }
.legal-label {
  font-family: var(--font-mono); font-size: 0.7rem; letter-spacing: 0.15em;
  text-transform: uppercase; color: var(--accent); display: block; margin-bottom: 0.5rem;
}
.legal-title {
  font-family: var(--font-serif); font-size: 2rem; font-weight: 700; margin-bottom: 0.5rem;
}
.legal-meta { font-size: 0.78rem; color: var(--text-muted); margin-bottom: 0; }
.legal-body { max-width: 720px; }
.legal-section { margin-bottom: 2rem; }
.legal-section h2 {
  font-family: var(--font-mono); font-size: 0.78rem; letter-spacing: 0.1em;
  text-transform: uppercase; color: var(--accent); margin-bottom: 0.8rem;
}
.legal-section p { color: var(--text-muted); font-size: 0.92rem; line-height: 1.75; margin-bottom: 0.8rem; }
.legal-section ul { list-style: none; padding: 0; margin: 0 0 0.8rem; display: flex; flex-direction: column; gap: 0.4rem; }
.legal-section ul li {
  font-size: 0.92rem; color: var(--text-muted); line-height: 1.65; padding-left: 1.2rem; position: relative;
}
.legal-section ul li::before {
  content: '—'; position: absolute; left: 0; color: var(--accent); font-family: var(--font-mono);
}
.consent-notice {
  display: flex; gap: 1rem; align-items: flex-start;
  padding: 1rem 1.2rem; margin-bottom: 2rem;
  background: rgba(212,168,67,0.06); border: 1px solid var(--accent-dim);
  border-radius: 6px;
}
.notice-icon {
  font-family: var(--font-mono); font-size: 0.75rem; color: var(--accent);
  border: 1px solid var(--accent); border-radius: 50%; width: 20px; height: 20px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0; margin-top: 2px;
  line-height: 1;
}
.consent-notice p { color: var(--text-muted); font-size: 0.88rem; line-height: 1.65; margin: 0; }
</style>
