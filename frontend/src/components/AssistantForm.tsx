import { useForm } from "@tanstack/react-form";
import { FormField } from "@/components/FormField";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import type { Category } from "@/lib/api";
import { fieldErrorMessage } from "@/lib/forms/fieldError";
import { assistantSchema, type AssistantFormValues } from "@/lib/forms/schemas/assistantSchema";

export type { AssistantFormValues };
export type AssistantFormOut = {
  name: string;
  description: string;
  categoryId: string;
  model: string;
  systemPrompt: string;
  exampleUserPrompt: string | null;
  isActive: boolean;
};

type Props = {
  mode: "create" | "edit";
  defaultValues: AssistantFormValues;
  categories: Category[];
  submitLabel: string;
  onSubmit: (values: AssistantFormOut) => Promise<void>;
  onCancel: () => void;
};

export function AssistantForm({
  mode,
  defaultValues,
  categories,
  submitLabel,
  onSubmit,
  onCancel,
}: Props) {
  const form = useForm({
    defaultValues,
    validators: { onChange: assistantSchema },
    onSubmit: async ({ value }) => {
      const parsed = assistantSchema.parse(value);
      await onSubmit({
        name: parsed.name,
        description: parsed.description,
        categoryId: parsed.categoryId,
        model: parsed.model,
        systemPrompt: parsed.systemPrompt,
        exampleUserPrompt:
          parsed.exampleUserPrompt === "" ? null : parsed.exampleUserPrompt,
        isActive: parsed.isActive,
      });
    },
  });

  const categoryLabels: Record<string, string> = {};
  for (const c of categories) categoryLabels[c.id] = c.name;

  return (
    <form
      className="flex max-w-2xl flex-col gap-5"
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <form.Field name="name">
        {(field) => (
          <FormField label="Название" error={fieldErrorMessage(field.state.meta.errors)}>
            {(p) => (
              <Input
                {...p}
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="Например, Повар"
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="description">
        {(field) => (
          <FormField label="Описание" error={fieldErrorMessage(field.state.meta.errors)}>
            {(p) => (
              <Textarea
                {...p}
                rows={3}
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="Что делает ассистент"
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="categoryId">
        {(field) => (
          <FormField label="Категория" error={fieldErrorMessage(field.state.meta.errors)}>
            {(p) => (
              <Select
                value={field.state.value === "" ? undefined : field.state.value}
                onValueChange={(v) => field.handleChange(v ?? "")}
              >
                <SelectTrigger id={p.id} aria-invalid={p["aria-invalid"]}>
                  <SelectValue placeholder="Выберите категорию" items={categoryLabels} />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="model">
        {(field) => (
          <FormField
            label="Модель"
            description="Идентификатор модели, например mock-smart"
            error={fieldErrorMessage(field.state.meta.errors)}
          >
            {(p) => (
              <Input
                {...p}
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="mock-smart"
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="systemPrompt">
        {(field) => (
          <FormField
            label="Системный промпт"
            description="Инструкции для модели, скрыты от обычных пользователей"
            error={fieldErrorMessage(field.state.meta.errors)}
          >
            {(p) => (
              <Textarea
                {...p}
                rows={6}
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="Опиши правила поведения ассистента..."
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="exampleUserPrompt">
        {(field) => (
          <FormField
            label="Пример пользовательского промпта"
            description="Необязательно. Показывается как placeholder на странице ассистента"
            error={fieldErrorMessage(field.state.meta.errors)}
          >
            {(p) => (
              <Textarea
                {...p}
                rows={2}
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                placeholder="курица, рис, томаты, сыр"
              />
            )}
          </FormField>
        )}
      </form.Field>

      <form.Field name="isActive">
        {(field) => (
          <div className="flex items-center gap-3">
            <Switch
              id="assistant-is-active"
              checked={field.state.value}
              onCheckedChange={(v) => field.handleChange(v)}
            />
            <Label htmlFor="assistant-is-active">
              {mode === "create" ? "Активен сразу после создания" : "Активен"}
            </Label>
          </div>
        )}
      </form.Field>

      <form.Subscribe
        selector={(s) => ({ canSubmit: s.canSubmit, isSubmitting: s.isSubmitting })}
      >
        {({ canSubmit, isSubmitting }) => (
          <div className="flex gap-3">
            <Button
              type="submit"
              variant="black"
              size="lg"
              disabled={!canSubmit || isSubmitting}
            >
              {isSubmitting ? "Сохраняем..." : submitLabel}
            </Button>
            <Button
              type="button"
              variant="soft"
              size="lg"
              onClick={onCancel}
              disabled={isSubmitting}
            >
              Отмена
            </Button>
          </div>
        )}
      </form.Subscribe>
    </form>
  );
}
