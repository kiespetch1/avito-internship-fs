import { useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert.tsx";
import { Avatar, AvatarFallback } from "@/components/ui/avatar.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { Button } from "@/components/ui/button.tsx";
import { Card, CardTitle } from "@/components/ui/card.tsx";
import { Checkbox, CheckboxGroup } from "@/components/ui/checkbox.tsx";
import { Tag } from "@/components/ui/tag.tsx";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog.tsx";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination.tsx";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group.tsx";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select.tsx";
import { Separator } from "@/components/ui/separator.tsx";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet.tsx";
import { Skeleton } from "@/components/ui/skeleton.tsx";
import { Switch } from "@/components/ui/switch.tsx";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table.tsx";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs.tsx";
import { Textarea } from "@/components/ui/textarea.tsx";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb.tsx";
import {
  AlertCircle,
  Bell,
  CheckCircle2,
  ChevronDown,
  Heart,
  Loader2,
  LogOut,
  MessageCircle,
  Search,
  Settings,
  ShoppingCart,
  User,
} from "lucide-react";

type RunStatus = "pending" | "success" | "failed";

function StatusBadge({ status }: { status: RunStatus }) {
  if (status === "success") return <Tag variant="success">Успешно</Tag>;
  if (status === "failed") return <Tag variant="destructive">Ошибка</Tag>;
  return <Tag variant="secondary">Ожидание</Tag>;
}

function Section({
  title,
  hint,
  children,
}: {
  title: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-xl font-bold tracking-tight">{title}</h2>
        {hint && <p className="mt-1 text-sm text-muted-foreground">{hint}</p>}
      </div>
      {children}
      <Separator />
    </section>
  );
}

function ColorSwatch({
  name,
  value,
  className,
}: {
  name: string;
  value: string;
  className: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <div className={`h-16 w-full rounded-lg border ${className}`} />
      <div className="text-xs font-semibold">{name}</div>
      <div className="text-xs text-muted-foreground">{value}</div>
    </div>
  );
}

export function UiKitPage() {
  const [showInactive, setShowInactive] = useState(false);
  const [role, setRole] = useState("user");
  const [activeChip, setActiveChip] = useState("all");

  return (
    <div className="mx-auto max-w-5xl space-y-10 px-6 py-12">
      <div>
        <p className="text-sm font-semibold text-primary">UI Kit · Авито-стиль</p>
        <h1 className="mt-1 text-4xl font-extrabold tracking-tight">
          Компоненты каталога
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Цветовая гамма, кнопки и бейджи в духе Авито
        </p>
      </div>

      <Separator />

      {/* ── Цвета ── */}
      <Section title="Цветовая палитра" hint="Дизайн-токены через CSS-переменные">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 lg:grid-cols-6">
          <ColorSwatch name="Primary" value="#00AAFF" className="bg-primary" />
          <ColorSwatch
            name="Avito Black"
            value="#0A0A0A"
            className="bg-avito-black"
          />
          <ColorSwatch
            name="Purple"
            value="#965EEB"
            className="bg-avito-purple"
          />
          <ColorSwatch
            name="Destructive"
            value="#FF4053"
            className="bg-destructive"
          />
          <ColorSwatch
            name="Success"
            value="#04C270"
            className="bg-avito-success"
          />
          <ColorSwatch
            name="Warning"
            value="#FFB02E"
            className="bg-avito-warning"
          />
          <ColorSwatch
            name="Secondary"
            value="#F4F5F6"
            className="bg-secondary"
          />
          <ColorSwatch name="Accent" value="#EFF0F1" className="bg-accent" />
          <ColorSwatch
            name="Info BG"
            value="#E7F4FE"
            className="bg-avito-info-bg"
          />
          <ColorSwatch name="Border" value="#E4E6E8" className="bg-border" />
          <ColorSwatch
            name="Muted text"
            value="#8C8C8C"
            className="bg-muted-foreground"
          />
          <ColorSwatch
            name="Foreground"
            value="#0A0A0A"
            className="bg-foreground"
          />
        </div>
      </Section>

      {/* ── Кнопки ── */}
      <Section title="Кнопки" hint="Все варианты + размеры">
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-3">
            <Button>Найти</Button>
            <Button variant="black">Все категории</Button>
            <Button variant="purple">Купить с доставкой</Button>
            <Button variant="destructive">Удалить</Button>
            <Button variant="soft">Показать телефон</Button>
            <Button variant="outline">Редактировать</Button>
            <Button variant="ghost">Отмена</Button>
            <Button variant="link">Об Авито Доставке</Button>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Button size="lg">Запустить ассистента</Button>
            <Button>Стандарт</Button>
            <Button size="sm">Малая</Button>
            <Button size="xs">XS</Button>
            <Button disabled>
              <Loader2 className="animate-spin" />
              Запуск...
            </Button>
            <Button size="icon" variant="soft">
              <Heart />
            </Button>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {(
              [
                { id: "all", label: "Все 15" },
                { id: "biz", label: "Бизнес360 1" },
                { id: "apt", label: "Квартиры 7" },
                { id: "pc", label: "Товары для компьютера 5" },
              ] as const
            ).map((chip) => (
              <Button
                key={chip.id}
                size="pill"
                variant={activeChip === chip.id ? "black" : "soft"}
                onClick={() => setActiveChip(chip.id)}
              >
                {chip.label}
              </Button>
            ))}
          </div>
        </div>
      </Section>

      {/* ── Tag ── */}
      <Section title="Tag" hint="Pill-форма, для категорий, меток «Новое», статусов запусков">
        <div className="space-y-3">
          <div className="flex flex-wrap gap-2">
            <Tag>Primary</Tag>
            <Tag variant="black">Все 15</Tag>
            <Tag variant="purple">Promo</Tag>
            <Tag variant="destructive">Новое</Tag>
            <Tag variant="success">Активен</Tag>
            <Tag variant="secondary">Б/у</Tag>
            <Tag variant="outline">gpt-4o-mini</Tag>
          </div>
          <div className="flex flex-wrap gap-2">
            <StatusBadge status="pending" />
            <StatusBadge status="success" />
            <StatusBadge status="failed" />
          </div>
        </div>
      </Section>

      {/* ── Badge ── */}
      <Section title="Badge" hint="Умеренное скругление, для info-меток и характеристик">
        <div className="flex flex-wrap gap-2">
          <Badge>Надёжный партнёр</Badge>
          <Badge>3 продажи с Авито Доставкой</Badge>
          <Badge>Документы проверены</Badge>
          <Badge>Эковклад: −2,24 тонн CO₂</Badge>
          <Badge variant="default">Состояние: Б/у</Badge>
          <Badge variant="success">Успешный запуск</Badge>
          <Badge variant="black">Premium</Badge>
          <Badge variant="destructive">Жалобы: 2</Badge>
          <Badge variant="outline">Категория: Еда</Badge>
        </div>
      </Section>

      {/* ── Шапка а-ля Авито ── */}
      <Section title="Header — шапка приложения">
        <div className="flex items-center gap-4 rounded-2xl border bg-background p-4">
          <div className="flex items-center gap-2">
            <div className="grid h-7 w-7 grid-cols-2 gap-0.5">
              <span className="rounded-full bg-primary" />
              <span className="rounded-full bg-avito-purple" />
              <span className="rounded-full bg-avito-success" />
              <span className="rounded-full bg-destructive" />
            </div>
            <span className="text-lg font-extrabold tracking-tight">
              AI-каталог
            </span>
          </div>
          <Button variant="default" size="pill">
            <Search />
            Все категории
          </Button>
          <div className="relative flex-1">
            <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="h-10 rounded-full bg-secondary pl-9"
              placeholder="Поиск ассистентов"
            />
          </div>
          <Button size="icon" variant="ghost">
            <Heart />
          </Button>
          <Button size="icon" variant="ghost">
            <Bell />
          </Button>
          <Button size="icon" variant="ghost">
            <MessageCircle />
          </Button>
          <Button size="icon" variant="ghost" className="relative">
            <ShoppingCart />
            <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-destructive" />
          </Button>
          <Avatar className="h-9 w-9">
            <AvatarFallback className="bg-secondary text-sm font-bold">
              ДТ
            </AvatarFallback>
          </Avatar>
        </div>
      </Section>

      {/* ── Сайдбар nav ── */}
      <Section title="Sidebar nav" hint="Стиль ссылок профиля Авито">
        <div className="grid gap-6 sm:grid-cols-[220px_1fr]">
          <nav className="space-y-1 text-[0.9375rem]">
            {(
              [
                { label: "Мои объявления" },
                { label: "Сравнение цен", badge: "Новое" },
                { label: "Заказы" },
                { label: "Мои отзывы" },
                { label: "Избранное", active: true },
                { label: "Портал призов" },
                { label: "Приглашайте друзей", danger: true },
                { label: "Гараж" },
              ] as ReadonlyArray<{
                label: string;
                active?: boolean;
                badge?: string;
                danger?: boolean;
              }>
            ).map((item) => (
              <a
                key={item.label}
                href="#"
                className={
                  item.active
                    ? "flex items-center gap-2 font-bold text-foreground"
                    : item.danger
                      ? "flex items-center gap-2 text-destructive hover:underline"
                      : "flex items-center gap-2 text-primary hover:underline"
                }
              >
                <span>{item.label}</span>
                {item.badge && (
                  <Tag variant="destructive" className="h-5 px-1.5">
                    {item.badge}
                  </Tag>
                )}
              </a>
            ))}
          </nav>
          <div className="rounded-2xl bg-secondary p-6">
            <p className="text-sm text-muted-foreground">
              Активный пункт — жирный чёрный, обычные — синие ссылки, опасные —
              красные.
            </p>
          </div>
        </div>
      </Section>

      {/* ── Карточка ассистента ── */}
      <Section title="Карточка ассистента" hint="Авито-стиль плитки">
        <div className="grid gap-4 sm:grid-cols-3">
          <Card className="overflow-hidden rounded-2xl border bg-secondary p-0 shadow-none">
            <div className="aspect-[4/3] bg-gradient-to-br from-primary/20 to-avito-purple/20" />
            <div className="space-y-2 p-4">
              <div className="flex items-center justify-between">
                <Badge>Еда</Badge>
                <Heart className="size-5 text-muted-foreground" />
              </div>
              <CardTitle className="text-base font-bold">Повар</CardTitle>
              <p className="text-xs text-muted-foreground">
                Составляет домашние рецепты из ваших ингредиентов
              </p>
              <div className="flex items-center justify-between pt-1">
                <span className="text-xs font-semibold text-muted-foreground">
                  gpt-4o-mini
                </span>
                <Tag variant="success">Активен</Tag>
              </div>
            </div>
          </Card>

          <Card className="overflow-hidden rounded-2xl border bg-secondary p-0 shadow-none">
            <div className="aspect-[4/3] bg-gradient-to-br from-avito-warning/30 to-destructive/20" />
            <div className="space-y-2 p-4">
              <div className="flex items-center justify-between">
                <Badge>Спорт</Badge>
                <Heart className="size-5 text-muted-foreground" />
              </div>
              <CardTitle className="text-base font-bold">Тренер</CardTitle>
              <p className="text-xs text-muted-foreground">
                Составляет план тренировок под цели
              </p>
              <div className="flex items-center justify-between pt-1">
                <span className="text-xs font-semibold text-muted-foreground">
                  gpt-4o
                </span>
                <Tag variant="destructive">Новое</Tag>
              </div>
            </div>
          </Card>

          <Card className="overflow-hidden rounded-2xl border bg-secondary p-0 opacity-60 shadow-none">
            <div className="aspect-[4/3] bg-gradient-to-br from-muted to-accent" />
            <div className="space-y-2 p-4">
              <div className="flex items-center justify-between">
                <Badge>Программирование</Badge>
                <Heart className="size-5 text-muted-foreground" />
              </div>
              <CardTitle className="text-base font-bold">Code-ревьюер</CardTitle>
              <p className="text-xs text-muted-foreground">
                Анализирует diff и предлагает улучшения
              </p>
              <div className="flex items-center justify-between pt-1">
                <span className="text-xs font-semibold text-muted-foreground">
                  gpt-4o-mini
                </span>
                <Tag variant="secondary">Неактивен</Tag>
              </div>
            </div>
          </Card>
        </div>
      </Section>

      {/* ── Alert ── */}
      <Section title="Alert — ошибки и успех">
        <Alert>
          <CheckCircle2 className="h-4 w-4" />
          <AlertTitle>Ассистент успешно запущен</AlertTitle>
          <AlertDescription>
            Ответ получен и сохранён в истории запусков.
          </AlertDescription>
        </Alert>
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Ошибка запуска</AlertTitle>
          <AlertDescription>
            LLM-провайдер вернул ошибку. Запуск сохранён со статусом «failed».
          </AlertDescription>
        </Alert>
      </Section>

      {/* ── Skeleton ── */}
      <Section title="Skeleton — состояние загрузки">
        <div className="grid gap-4 sm:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card
              key={i}
              className="overflow-hidden rounded-2xl border bg-secondary p-0 shadow-none"
            >
              <Skeleton className="aspect-[4/3] w-full rounded-none" />
              <div className="space-y-2 p-4">
                <Skeleton className="h-4 w-20" />
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-3 w-full" />
                <Skeleton className="h-3 w-2/3" />
              </div>
            </Card>
          ))}
        </div>
      </Section>

      {/* ── Форма запуска ── */}
      <Section title="Form — запуск ассистента" hint="Стиль форм Авито: серые поля, без бордеров, label сверху жирно">
        <div className="max-w-xl space-y-6">
          <div>
            <h3 className="text-2xl font-extrabold tracking-tight">Повар</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Составит рецепт из ваших ингредиентов
            </p>
          </div>
          <div className="space-y-2">
            <Label className="text-[0.9375rem] font-bold">Ваш запрос</Label>
            <Textarea
              placeholder="Введите список ингредиентов..."
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              Пример: курица, лаваш, огурцы, томаты, соус
            </p>
          </div>
          <div className="flex gap-3">
            <Button variant="black" size="lg">
              Запустить
            </Button>
            <Button variant="soft" size="lg">
              Очистить
            </Button>
          </div>
        </div>
      </Section>

      {/* ── Форма создания ── */}
      <Section title="Form — создание ассистента (admin)">
        <div className="max-w-xl space-y-5">
          <div className="space-y-2">
            <Label className="text-[0.9375rem] font-bold">Название</Label>
            <Input placeholder="Повар" />
          </div>
          <div className="space-y-2">
            <Label className="text-[0.9375rem] font-bold">Описание</Label>
            <Textarea placeholder="Описание ассистента..." rows={2} />
          </div>
          <div className="space-y-2">
            <Label className="text-[0.9375rem] font-bold">Категория</Label>
            <Select>
              <SelectTrigger>
                <SelectValue placeholder="Выберите категорию" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="food">Еда</SelectItem>
                <SelectItem value="sport">Спорт</SelectItem>
                <SelectItem value="code">Программирование</SelectItem>
                <SelectItem value="travel">Путешествия</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label className="text-[0.9375rem] font-bold">Модель</Label>
            <Select>
              <SelectTrigger>
                <SelectValue placeholder="Выберите модель" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="gpt-4o-mini">gpt-4o-mini</SelectItem>
                <SelectItem value="gpt-4o">gpt-4o</SelectItem>
                <SelectItem value="claude-haiku-4-5">claude-haiku-4-5</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label className="text-[0.9375rem] font-bold">Системный промпт</Label>
            <Textarea
              placeholder="Инструкция для ассистента..."
              rows={4}
            />
          </div>
          <div className="flex items-center gap-3 pt-2">
            <Switch id="asst-active" defaultChecked />
            <Label htmlFor="asst-active" className="text-[0.9375rem] font-semibold">
              Ассистент активен
            </Label>
          </div>
          <div className="flex gap-3 pt-4">
            <Button variant="black" size="lg">
              Создать
            </Button>
            <Button variant="soft" size="lg">
              Сохранить и выйти
            </Button>
          </div>
        </div>
      </Section>

      {/* ── Настройки (демо switch + select + radio в стиле Авито) ── */}
      <Section title="Настройки — switch / select / radio">
        <div className="max-w-2xl space-y-8">
          <div className="space-y-4">
            <h3 className="text-xl font-extrabold">Звонки</h3>
            <p className="text-sm text-muted-foreground">
              Включаются в каждом объявлении отдельно.
            </p>
            <div className="space-y-2">
              <Label className="text-[0.9375rem] font-bold">Когда принимать</Label>
              <div className="flex flex-wrap items-center gap-4">
                <Select>
                  <SelectTrigger className="w-40">
                    <SelectValue placeholder="С 09:00" />
                  </SelectTrigger>
                  <SelectContent>
                    {Array.from({ length: 12 }, (_, i) => {
                      const h = i.toString().padStart(2, "0");
                      return (
                        <SelectItem key={h} value={`${h}:00`}>
                          {h}:00
                        </SelectItem>
                      );
                    })}
                  </SelectContent>
                </Select>
                <Select>
                  <SelectTrigger className="w-40">
                    <SelectValue placeholder="До 23:00" />
                  </SelectTrigger>
                  <SelectContent>
                    {Array.from({ length: 12 }, (_, i) => {
                      const h = (i + 12).toString().padStart(2, "0");
                      return (
                        <SelectItem key={h} value={`${h}:00`}>
                          {h}:00
                        </SelectItem>
                      );
                    })}
                  </SelectContent>
                </Select>
                <div className="flex items-center gap-2">
                  <Switch />
                  <Label className="text-[0.9375rem]">В любое время</Label>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3 pt-2">
              <Switch defaultChecked />
              <Label className="text-[0.9375rem] font-semibold">
                Звонки через Авито
              </Label>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <h3 className="text-xl font-extrabold">Способ связи</h3>
            <RadioGroup defaultValue="both">
              <div className="flex items-center gap-3">
                <RadioGroupItem value="both" id="contact-both" />
                <Label htmlFor="contact-both" className="text-[0.9375rem]">
                  Звонки и сообщения
                </Label>
              </div>
              <div className="flex items-center gap-3">
                <RadioGroupItem value="calls" id="contact-calls" />
                <Label htmlFor="contact-calls" className="text-[0.9375rem]">
                  Только звонки
                </Label>
              </div>
              <div className="flex items-center gap-3">
                <RadioGroupItem value="messages" id="contact-msg" />
                <Label htmlFor="contact-msg" className="text-[0.9375rem]">
                  Только сообщения
                </Label>
              </div>
            </RadioGroup>
          </div>

          <Separator />

          <div className="flex gap-3">
            <Button variant="black" size="lg">
              Разместить
            </Button>
            <Button variant="soft" size="lg">
              Сохранить и выйти
            </Button>
          </div>
        </div>
      </Section>

      {/* ── Checkbox / мультивыбор ── */}
      <Section title="Checkbox — мультивыбор" hint="График работы, тип занятости и т.п.">
        <div className="max-w-2xl space-y-8">
          <div className="space-y-4">
            <h3 className="text-xl font-extrabold">Тип занятости</h3>
            <CheckboxGroup
              defaultValue={["main"]}
              className="grid-cols-2"
            >
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="main" />
                Основная работа
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="parttime" />
                Подработка
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="project" />
                Проектная работа
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="internship" />
                Стажировка
              </Label>
            </CheckboxGroup>
          </div>

          <Separator />

          <div className="space-y-4">
            <h3 className="text-xl font-extrabold">График работы</h3>
            <CheckboxGroup className="grid-cols-2">
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="shift" />
                Вахтовый метод
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="flexible" />
                Свободный график
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="parttime-day" />
                Неполный день
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox name="shift-rotation" />
                Сменный график
              </Label>
            </CheckboxGroup>
          </div>

          <Separator />

          <div className="space-y-4">
            <h3 className="text-xl font-extrabold">Состояния</h3>
            <div className="flex items-center gap-6">
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox defaultChecked />
                Отмечено
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox />
                Не отмечено
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem]">
                <Checkbox indeterminate />
                Частично
              </Label>
              <Label className="flex items-center gap-3 text-[0.9375rem] opacity-50">
                <Checkbox disabled />
                Disabled
              </Label>
            </div>
          </div>
        </div>
      </Section>

      {/* ── Фильтры каталога ── */}
      <Section title="Фильтры каталога ассистентов">
        <div className="flex flex-wrap gap-3">
          <div className="relative w-64">
            <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="pl-9"
              placeholder="Поиск по названию или описанию"
            />
          </div>
          <Select>
            <SelectTrigger className="w-44">
              <SelectValue placeholder="Категория" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Все категории</SelectItem>
              <SelectItem value="food">Еда</SelectItem>
              <SelectItem value="sport">Спорт</SelectItem>
            </SelectContent>
          </Select>
          <div className="flex items-center gap-2">
            <Switch
              id="filter-inactive"
              checked={showInactive}
              onCheckedChange={setShowInactive}
            />
            <Label htmlFor="filter-inactive">Показать неактивных</Label>
          </div>
        </div>
      </Section>

      {/* ── Login ── */}
      <Section title="Login — выбор роли" hint="Radio как карточки-row (стиль Авито)">
        <div className="max-w-md space-y-6">
          <div>
            <h3 className="text-2xl font-extrabold tracking-tight">Войти</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Выберите роль для тестового входа
            </p>
          </div>
          <RadioGroup
            value={role}
            onValueChange={setRole}
            className="gap-2"
          >
            <Label
              htmlFor="role-user"
              className="flex cursor-pointer items-start gap-3 rounded-2xl bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
            >
              <div className="flex-1 space-y-1">
                <div className="text-[0.9375rem] font-bold">Пользователь</div>
                <p className="text-sm text-muted-foreground">
                  Просмотр каталога и запуск ассистентов
                </p>
              </div>
              <RadioGroupItem value="user" id="role-user" />
            </Label>
            <Label
              htmlFor="role-admin"
              className="flex cursor-pointer items-start gap-3 rounded-2xl bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
            >
              <div className="flex-1 space-y-1">
                <div className="text-[0.9375rem] font-bold">Администратор</div>
                <p className="text-sm text-muted-foreground">
                  Управление категориями и ассистентами + история всех запусков
                </p>
              </div>
              <RadioGroupItem value="admin" id="role-admin" />
            </Label>
          </RadioGroup>
          <Button variant="black" size="lg" className="w-full">
            Войти как {role === "admin" ? "Admin" : "User"}
          </Button>
        </div>
      </Section>

      {/* ── Таблица запусков ── */}
      <Section title="Table — история запусков">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Ассистент</TableHead>
              <TableHead>Запрос</TableHead>
              <TableHead>Статус</TableHead>
              <TableHead>Дата</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(
              [
                {
                  name: "Повар",
                  prompt: "курица, рис, томаты",
                  status: "success",
                  date: "12.05.2026",
                },
                {
                  name: "Тренер",
                  prompt: "похудеть за месяц",
                  status: "failed",
                  date: "11.05.2026",
                },
                {
                  name: "Повар",
                  prompt: "яйца, сыр",
                  status: "pending",
                  date: "12.05.2026",
                },
              ] as const
            ).map((row, i) => (
              <TableRow key={i}>
                <TableCell className="font-semibold">{row.name}</TableCell>
                <TableCell className="max-w-40 truncate text-muted-foreground">
                  {row.prompt}
                </TableCell>
                <TableCell>
                  <StatusBadge status={row.status} />
                </TableCell>
                <TableCell>{row.date}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Section>

      {/* ── Tabs ── */}
      <Section title="Tabs — стиль Авито" hint="Жирный текст, тонкая чёрная линия под активным; счётчик опционален">
        <div className="space-y-8">
          <Tabs defaultValue="catalog">
            <TabsList>
              <TabsTrigger value="catalog" count={42}>
                Каталог
              </TabsTrigger>
              <TabsTrigger value="runs" count={7}>
                Мои запуски
              </TabsTrigger>
              <TabsTrigger value="favs" count={3}>
                Избранное
              </TabsTrigger>
            </TabsList>
            <TabsContent value="catalog" className="pt-4">
              <p className="text-sm text-muted-foreground">Список ассистентов...</p>
            </TabsContent>
            <TabsContent value="runs" className="pt-4">
              <p className="text-sm text-muted-foreground">История запусков...</p>
            </TabsContent>
            <TabsContent value="favs" className="pt-4">
              <p className="text-sm text-muted-foreground">Избранные ассистенты...</p>
            </TabsContent>
          </Tabs>

          <Tabs defaultValue="overview">
            <TabsList>
              <TabsTrigger value="overview">Обзор</TabsTrigger>
              <TabsTrigger value="settings">Настройки</TabsTrigger>
              <TabsTrigger value="logs">Логи</TabsTrigger>
            </TabsList>
            <TabsContent value="overview" className="pt-4">
              <p className="text-sm text-muted-foreground">Без счётчиков.</p>
            </TabsContent>
            <TabsContent value="settings" className="pt-4">
              <p className="text-sm text-muted-foreground">Настройки ассистента.</p>
            </TabsContent>
            <TabsContent value="logs" className="pt-4">
              <p className="text-sm text-muted-foreground">Журнал запусков.</p>
            </TabsContent>
          </Tabs>
        </div>
      </Section>

      {/* ── Pagination ── */}
      <Section title="Pagination">
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious href="#" />
            </PaginationItem>
            <PaginationItem>
              <PaginationLink href="#">1</PaginationLink>
            </PaginationItem>
            <PaginationItem>
              <PaginationLink href="#" isActive>
                2
              </PaginationLink>
            </PaginationItem>
            <PaginationItem>
              <PaginationLink href="#">3</PaginationLink>
            </PaginationItem>
            <PaginationItem>
              <PaginationEllipsis />
            </PaginationItem>
            <PaginationItem>
              <PaginationNext href="#" />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </Section>

      {/* ── Breadcrumb ── */}
      <Section title="Breadcrumb — крошки" hint="Серый текст, разделитель — chevron">
        <div className="space-y-3">
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink href="#">Главная</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbLink href="#">Ассистенты</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbLink href="#">Еда</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>Повар</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink href="#">Работа</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>Резюме</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </Section>

      {/* ── DropdownMenu профиля ── */}
      <Section title="Dropdown профиля">
        <div className="flex justify-end rounded-2xl border bg-background p-4">
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button variant="ghost" className="gap-2">
                  <Avatar className="h-7 w-7">
                    <AvatarFallback className="bg-secondary text-xs font-bold">
                      ДТ
                    </AvatarFallback>
                  </Avatar>
                  <span className="text-sm font-semibold">Данил</span>
                  <ChevronDown className="h-4 w-4" />
                </Button>
              }
            />
            <DropdownMenuContent align="end">
              <DropdownMenuItem>
                <User className="mr-2 h-4 w-4" />
                Профиль
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Heart className="mr-2 h-4 w-4" />
                Избранное
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Settings className="mr-2 h-4 w-4" />
                Настройки
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-destructive">
                <LogOut className="mr-2 h-4 w-4" />
                Выйти
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </Section>

      {/* ── Dialog ── */}
      <Section title="Dialog — модалки в стиле Авито" hint="Success-alert и confirmation с выбором">
        <div className="flex flex-wrap gap-3">
          {/* Success-alert: иконка-кружок, заголовок, текст, одна чёрная кнопка */}
          <Dialog>
            <DialogTrigger
              render={<Button variant="soft">Success-alert</Button>}
            />
            <DialogContent>
              <div className="flex size-12 items-center justify-center rounded-2xl bg-avito-success/20">
                <CheckCircle2 className="size-7 text-avito-success" />
              </div>
              <DialogHeader>
                <DialogTitle>Ассистент успешно запущен</DialogTitle>
                <DialogDescription>
                  Ответ получен и сохранён в истории запусков.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button variant="black" size="lg">
                  Хорошо
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Confirmation с выбором: иллюстрация сверху, radio-карточки, две кнопки */}
          <Dialog>
            <DialogTrigger
              render={<Button variant="soft">Confirmation</Button>}
            />
            <DialogContent>
              <div className="flex h-32 items-center justify-center rounded-2xl bg-avito-purple/15">
                <div className="size-20 rounded-2xl bg-gradient-to-br from-avito-purple/50 to-primary/40" />
              </div>
              <DialogHeader>
                <DialogTitle>Что пойдёт на проверку?</DialogTitle>
              </DialogHeader>
              <RadioGroup defaultValue="medbook" className="gap-2">
                <Label
                  htmlFor="r-medbook"
                  className="flex cursor-pointer items-start gap-3 rounded-2xl bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
                >
                  <div className="flex-1 space-y-1">
                    <div className="text-[0.9375rem] font-bold">Медкнижка</div>
                    <p className="text-sm text-muted-foreground">
                      Нужны фотографии трёх страниц
                    </p>
                  </div>
                  <RadioGroupItem value="medbook" id="r-medbook" />
                </Label>
                <Label
                  htmlFor="r-cert"
                  className="flex cursor-pointer items-start gap-3 rounded-2xl bg-secondary p-4 has-data-checked:ring-2 has-data-checked:ring-avito-black"
                >
                  <div className="flex-1 space-y-1">
                    <div className="text-[0.9375rem] font-bold">Справка</div>
                    <p className="text-sm text-muted-foreground">
                      Получите её там, где оформляете медкнижку. Действует 21
                      день
                    </p>
                  </div>
                  <RadioGroupItem value="cert" id="r-cert" />
                </Label>
              </RadioGroup>
              <DialogFooter>
                <Button variant="black" size="lg">
                  Продолжить
                </Button>
                <Button variant="soft" size="lg">
                  Как оформить медкнижку
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Простая модалка с длинным текстом — наш юзкейс «полный ответ ассистента» */}
          <Dialog>
            <DialogTrigger
              render={<Button variant="soft">Полный ответ ассистента</Button>}
            />
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Ответ ассистента «Повар»</DialogTitle>
                <DialogDescription>
                  Запрос: курица, лаваш, огурцы, томаты, соус
                </DialogDescription>
              </DialogHeader>
              <div className="rounded-2xl bg-secondary p-4 text-[0.9375rem] leading-relaxed">
                Домашняя шаверма за 20 минут: нарежь курицу полосками и обжарь
                до золотистой корочки, огурцы и томаты порежь тонкими ломтиками,
                смажь лаваш соусом, выложи овощи и курицу, плотно заверни и
                подрумянь на сухой сковороде по минуте с каждой стороны.
              </div>
              <DialogFooter>
                <Button variant="black" size="lg">
                  Закрыть
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </Section>

      {/* ── Sheet ── */}
      <Section title="Sheet — мобильные фильтры">
        <Sheet>
          <SheetTrigger
            render={<Button variant="soft">Открыть фильтры</Button>}
          />
          <SheetContent>
            <SheetHeader>
              <SheetTitle>Фильтры</SheetTitle>
              <SheetDescription>
                Настройте отображение ассистентов
              </SheetDescription>
            </SheetHeader>
            <div className="mt-6 space-y-4 px-4">
              <div className="space-y-2">
                <Label>Категория</Label>
                <Select>
                  <SelectTrigger>
                    <SelectValue placeholder="Все категории" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="food">Еда</SelectItem>
                    <SelectItem value="sport">Спорт</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <Switch id="sheet-inactive" />
                <Label htmlFor="sheet-inactive">Показать неактивных</Label>
              </div>
              <Button size="lg" className="w-full">
                Применить
              </Button>
            </div>
          </SheetContent>
        </Sheet>
      </Section>
    </div>
  );
}
