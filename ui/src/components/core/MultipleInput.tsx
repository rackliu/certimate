﻿import { forwardRef, useImperativeHandle, useMemo, useRef, type ChangeEvent } from "react";
import { useTranslation } from "react-i18next";
import { useControllableValue } from "ahooks";
import { Button, Input, Space, type InputRef, type InputProps } from "antd";
import {
  ArrowDownOutlined as ArrowDownOutlinedIcon,
  ArrowUpOutlined as ArrowUpOutlinedIcon,
  MinusOutlined as MinusOutlinedIcon,
  PlusOutlined as PlusOutlinedIcon,
} from "@ant-design/icons";
import { produce } from "immer";

export type MultipleInputProps = Omit<InputProps, "count" | "defaultValue" | "showCount" | "value" | "onChange" | "onPressEnter" | "onClear"> & {
  allowClear?: boolean;
  defaultValue?: string[];
  maxCount?: number;
  minCount?: number;
  showSortButton?: boolean;
  value?: string[];
  onChange?: (value: string[]) => void;
  onValueChange?: (index: number, element: string) => void;
  onValueCreate?: (index: number) => void;
  onValueRemove?: (index: number) => void;
  onValueSort?: (oldIndex: number, newIndex: number) => void;
};

const MultipleInput = ({
  allowClear = false,
  disabled,
  maxCount,
  minCount,
  showSortButton = true,
  onValueChange,
  onValueCreate,
  onValueSort,
  onValueRemove,
  ...props
}: MultipleInputProps) => {
  const { t } = useTranslation();

  const itemRefs = useRef<MultipleInputItemInstance[]>([]);

  const [value, setValue] = useControllableValue<string[]>(props, {
    valuePropName: "value",
    defaultValue: [],
    defaultValuePropName: "defaultValue",
    trigger: "onChange",
  });

  const handleCreate = () => {
    const newValue = produce(value, (draft) => {
      draft.push("");
    });
    setValue(newValue);
    setTimeout(() => itemRefs.current[newValue.length - 1]?.focus(), 0);

    onValueCreate?.(newValue.length - 1);
  };

  const handleChange = (index: number, element: string) => {
    const newValue = produce(value, (draft) => {
      draft[index] = element;
    });
    setValue(newValue);

    onValueChange?.(index, element);
  };

  const handleInputBlur = (index: number) => {
    if (!allowClear && !value[index]) {
      const newValue = produce(value, (draft) => {
        draft.splice(index, 1);
      });
      setValue(newValue);
    }
  };

  const handleClickUp = (index: number) => {
    if (index === 0) {
      return;
    }

    const newValue = produce(value, (draft) => {
      const temp = draft[index - 1];
      draft[index - 1] = draft[index];
      draft[index] = temp;
    });
    setValue(newValue);

    onValueSort?.(index, index - 1);
  };

  const handleClickDown = (index: number) => {
    if (index === value.length - 1) {
      return;
    }

    const newValue = produce(value, (draft) => {
      const temp = draft[index + 1];
      draft[index + 1] = draft[index];
      draft[index] = temp;
    });
    setValue(newValue);

    onValueSort?.(index, index + 1);
  };

  const handleClickAdd = (index: number) => {
    const newValue = produce(value, (draft) => {
      draft.splice(index + 1, 0, "");
    });
    setValue(newValue);
    setTimeout(() => itemRefs.current[index + 1]?.focus(), 0);

    onValueCreate?.(index + 1);
  };

  const handleClickRemove = (index: number) => {
    const newValue = produce(value, (draft) => {
      draft.splice(index, 1);
    });
    setValue(newValue);

    onValueRemove?.(index);
  };

  return value == null || value.length === 0 ? (
    <Button block color="primary" disabled={disabled || maxCount === 0} size={props.size} variant="dashed" onClick={handleCreate}>
      {t("common.button.add")}
    </Button>
  ) : (
    <Space className="w-full" direction="vertical" size="small">
      {Array.from(value).map((element, index) => {
        const allowUp = index > 0;
        const allowDown = index < value.length - 1;
        const allowRemove = minCount == null || value.length > minCount;
        const allowAdd = maxCount == null || value.length < maxCount;

        return (
          <MultipleInputItem
            {...props}
            key={index}
            ref={(ref) => (itemRefs.current[index] = ref!)}
            allowAdd={allowAdd}
            allowClear={allowClear}
            allowDown={allowDown}
            allowRemove={allowRemove}
            allowUp={allowUp}
            disabled={disabled}
            defaultValue={undefined}
            showSortButton={showSortButton}
            value={element}
            onBlur={() => handleInputBlur(index)}
            onChange={(val) => handleChange(index, val)}
            onClickAdd={() => handleClickAdd(index)}
            onClickDown={() => handleClickDown(index)}
            onClickUp={() => handleClickUp(index)}
            onClickRemove={() => handleClickRemove(index)}
          />
        );
      })}
    </Space>
  );
};

type MultipleInputItemProps = Omit<
  MultipleInputProps,
  "defaultValue" | "maxCount" | "minCount" | "preset" | "value" | "onChange" | "onValueCreate" | "onValueRemove" | "onValueSort" | "onValueChange"
> & {
  allowAdd: boolean;
  allowRemove: boolean;
  allowUp: boolean;
  allowDown: boolean;
  defaultValue?: string;
  value?: string;
  onChange?: (value: string) => void;
  onClickAdd?: () => void;
  onClickDown?: () => void;
  onClickUp?: () => void;
  onClickRemove?: () => void;
};

type MultipleInputItemInstance = {
  focus: InputRef["focus"];
  blur: InputRef["blur"];
  select: InputRef["select"];
};

const MultipleInputItem = forwardRef<MultipleInputItemInstance, MultipleInputItemProps>(
  (
    {
      allowAdd,
      allowClear,
      allowDown,
      allowRemove,
      allowUp,
      disabled,
      showSortButton,
      size,
      onClickAdd,
      onClickDown,
      onClickUp,
      onClickRemove,
      ...props
    }: MultipleInputItemProps,
    ref
  ) => {
    const inputRef = useRef<InputRef>(null);

    const [value, setValue] = useControllableValue<string>(props, {
      valuePropName: "value",
      defaultValue: "",
      defaultValuePropName: "defaultValue",
      trigger: "onChange",
    });

    const upBtn = useMemo(() => {
      if (!showSortButton) return null;
      return <Button icon={<ArrowUpOutlinedIcon />} color="default" disabled={disabled || !allowUp} type="text" onClick={onClickUp} />;
    }, [allowUp, disabled, showSortButton, onClickUp]);
    const downBtn = useMemo(() => {
      if (!showSortButton) return null;
      return <Button icon={<ArrowDownOutlinedIcon />} color="default" disabled={disabled || !allowDown} type="text" onClick={onClickDown} />;
    }, [allowDown, disabled, showSortButton, onClickDown]);
    const removeBtn = useMemo(() => {
      return <Button icon={<MinusOutlinedIcon />} color="default" disabled={disabled || !allowRemove} type="text" onClick={onClickRemove} />;
    }, [allowRemove, disabled, onClickRemove]);
    const addBtn = useMemo(() => {
      return <Button icon={<PlusOutlinedIcon />} color="default" disabled={disabled || !allowAdd} type="text" onClick={onClickAdd} />;
    }, [allowAdd, disabled, onClickAdd]);

    const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
      setValue(e.target.value);
    };

    useImperativeHandle(ref, () => ({
      focus: (options) => {
        inputRef.current?.focus(options);
      },
      blur: () => {
        inputRef.current?.blur();
      },
      select: () => {
        inputRef.current?.select();
      },
    }));

    return (
      <div className="flex flex-nowrap items-center space-x-2">
        <div className="flex-grow">
          <Input
            {...props}
            ref={inputRef}
            className={undefined}
            style={undefined}
            allowClear={allowClear}
            defaultValue={undefined}
            value={value}
            onChange={handleInputChange}
          />
        </div>
        <Button.Group size={size}>
          {removeBtn}
          {upBtn}
          {downBtn}
          {addBtn}
        </Button.Group>
      </div>
    );
  }
);

export default MultipleInput;
