import { useForm } from "@tanstack/react-form";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { FormField } from "@/components/FormField";
import { Button } from "@/components/ui/button";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { showErrorToast } from "@/lib/api/errorMessage";
import { useCreateCategory } from "@/lib/api/queries/useCreateCategory";
import { fieldErrorMessage } from "@/lib/forms/fieldError";
import { categorySchema, type CategoryFormValues } from "@/lib/forms/schemas/categorySchema";

const defaults: CategoryFormValues = { name: "", description: "" };

export function NewCategoryPage() {
  const navigate = useNavigate();
  const createMutation = useCreateCategory();

  const form = useForm({
    defaultValues: defaults,
    validators: { onChange: categorySchema },
    onSubmit: async ({ value }) => {
      const parsed = categorySchema.parse(value);
      try {
        await createMutation.mutateAsync({
          name: parsed.name,
          description: parsed.description === "" ? null : parsed.description,
        });
        toast.success("Категория создана");
        navigate("/assistants");
      } catch (err) {
        showErrorToast(err, { fallback: "Не удалось создать категорию" });
      }
    },
  });

  return (
    <section className="flex flex-col gap-8">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/assistants">Каталог</BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Новая категория</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <header>
        <p className="text-sm font-semibold text-primary">Администрирование</p>
        <h1 className="mt-1 text-4xl font-extrabold tracking-tight">
          Новая категория
        </h1>
      </header>

      <form
        className="flex max-w-xl flex-col gap-5"
        onSubmit={(e) => {
          e.preventDefault();
          void form.handleSubmit();
        }}
      >
        <form.Field name="name">
          {(field) => (
            <FormField
              label="Название"
              error={fieldErrorMessage(field.state.meta.errors)}
            >
              {(p) => (
                <Input
                  {...p}
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Например, Еда"
                />
              )}
            </FormField>
          )}
        </form.Field>

        <form.Field name="description">
          {(field) => (
            <FormField
              label="Описание"
              description="Необязательно"
              error={fieldErrorMessage(field.state.meta.errors)}
            >
              {(p) => (
                <Textarea
                  {...p}
                  rows={3}
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Короткое описание категории"
                />
              )}
            </FormField>
          )}
        </form.Field>

        <form.Subscribe selector={(s) => ({ canSubmit: s.canSubmit, isSubmitting: s.isSubmitting })}>
          {({ canSubmit, isSubmitting }) => (
            <div className="flex gap-3">
              <Button
                type="submit"
                variant="black"
                size="lg"
                disabled={!canSubmit || isSubmitting}
              >
                {isSubmitting ? "Создаём..." : "Создать"}
              </Button>
              <Button
                type="button"
                variant="soft"
                size="lg"
                onClick={() => navigate("/assistants")}
                disabled={isSubmitting}
              >
                Отмена
              </Button>
            </div>
          )}
        </form.Subscribe>
      </form>
    </section>
  );
}
