package gejie

type CssSelector string

const ProductUrlExample = "https://articulo.mercadolibre.com.mx/MLM-1411526559-silla-gamer-reclinable-giratoria-ergonomica-super-comoda-_JM"
const exampleMercadoLibreXiaoMi15 = "https://listado.mercadolibre.com.pe/xiaomi-15"
const exampleMercadoLibreKeyboard = "https://listado.mercadolibre.com.pe/teclado-mecanico"
const exampleMeliSearchUrlCarbs = "https://listado.mercadolibre.com.mx/carburador-stihl"

const productImagesSelector CssSelector = ".ui-pdp-gallery__figure__image"
const productLinksSelector = ".ui-search-main--only-products div.poly-card__content > h3 > a"
const nameSelector CssSelector = "h1.ui-pdp-title"

const minimumSoldCountSelector CssSelector = "div.ui-pdp-header__subtitle > span.ui-pdp-subtitle"

const priceBoxSelector CssSelector = "#price > div > div.ui-pdp-price__main-container > div.ui-pdp-price__second-line > span > span"

const priceCurrencySelector CssSelector = priceBoxSelector + " .andes-money-amount__currency-symbol"
const priceAmountFractionSelector CssSelector = priceBoxSelector + " .andes-money-amount__fraction"
const priceAmountCentSelector CssSelector = priceBoxSelector + " .andes-money-amount__cents"

const reviewsContainerSelector CssSelector = "div.ui-pdp-header__info > a"
const reviewsRatingSelector CssSelector = "span.ui-pdp-review__rating"
const reviewsCountSelector CssSelector = "span.ui-pdp-review__amount"

const storeNameSelector CssSelector = "div.ui-seller-data-header__title-container > h2"
const storeUrlSelector CssSelector = "div.ui-seller-data-footer__container > a"
const storeLogoImageSelector CssSelector = "div.ui-seller-data__logo-image img"

const paginationNextButtonSelector CssSelector = "li.andes-pagination__button.andes-pagination__button--next > a"
